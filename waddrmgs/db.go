// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package waddrmgr

import (
	"encoding/binary"
	"errors"
	"fmt"
	chainhash "github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-wallet/walletdb"
	"time"
)

var (
	// LatestMgrVersion is the most recent manager version.
	LatestMgrVersion = getLatestVersion()

	// latestMgrVersion is the most recent manager version as a variable so
	// the tests can change it to force errors.
	latestMgrVersion = LatestMgrVersion
)

// ObtainUserInputFunc is a function that reads a user input and returns it as
// a byte stream. It is used to accept data required during upgrades, for e.g.
// wallet seed and private passphrase.
type ObtainUserInputFunc func() ([]byte, error)

// maybeConvertDbError converts the passed error to a ManagerError with an
// error code of ErrDatabase if it is not already a ManagerError.  This is
// useful for potential errors returned from managed transaction an other parts
// of the walletdb database.
func maybeConvertDbError(err error) error {
	// When the error is already a ManagerError, just return it.
	if _, ok := err.(ManagerError); ok {
		return err
	}

	return managerError(ErrDatabase, err.Error(), err)
}

// syncStatus represents a address synchronization status stored in the
// database.
type syncStatus uint8

// These constants define the various supported sync status types.
//
// NOTE: These are currently unused but are being defined for the possibility
// of supporting sync status on a per-address basis.
const (
	ssNone    syncStatus = 0 // not iota as they need to be stable for db
	ssPartial syncStatus = 1
	ssFull    syncStatus = 2
)

// addressType represents a type of address stored in the database.
type addressType uint8

// These constants define the various supported address types.
const (
	adtChain  addressType = 0
	adtImport addressType = 1 // not iota as they need to be stable for db
	adtScript addressType = 2
)

// accountType represents a type of address stored in the database.
type accountType uint8

// These constants define the various supported account types.
const (
	// accountDefault is the current "default" account type within the
	// database. This is an account that re-uses the key derivation schema
	// of BIP0044-like accounts.
	accountDefault accountType = 0 // not iota as they need to be stable
)

// dbAccountRow houses information stored about an account in the database.
type dbAccountRow struct {
	acctType accountType
	rawData  []byte // Varies based on account type field.
}

// dbDefaultAccountRow houses additional information stored about a default
// BIP0044-like account in the database.
type dbDefaultAccountRow struct {
	dbAccountRow
	pubKeyEncrypted   []byte
	privKeyEncrypted  []byte
	nextExternalIndex uint32
	nextInternalIndex uint32
	name              string
}

// dbAddressRow houses common information stored about an address in the
// database.
type dbAddressRow struct {
	addrType   addressType
	account    uint32
	addTime    uint64
	syncStatus syncStatus
	rawData    []byte // Varies based on address type field.
}

// dbChainAddressRow houses additional information stored about a chained
// address in the database.
type dbChainAddressRow struct {
	dbAddressRow
	branch uint32
	index  uint32
}

// dbImportedAddressRow houses additional information stored about an imported
// public key address in the database.
type dbImportedAddressRow struct {
	dbAddressRow
	encryptedPubKey  []byte
	encryptedPrivKey []byte
}

// dbImportedAddressRow houses additional information stored about a script
// address in the database.
type dbScriptAddressRow struct {
	dbAddressRow
	encryptedHash   []byte
	encryptedScript []byte
}

// Key names for various database fields.
var (
	// nullVall is null byte used as a flag value in a bucket entry
	nullVal = []byte{0}

	// Bucket names.

	// scopeSchemaBucket is the name of the bucket that maps a particular
	// manager scope to the type of addresses that should be derived for
	// particular branches during key derivation.
	scopeSchemaBucketName = []byte("scope-schema")

	// scopeBucketNme is the name of the top-level bucket within the
	// hierarchy. It maps: purpose || coinType to a new sub-bucket that
	// will house a scoped address manager. All buckets below are a child
	// of this bucket:
	//
	// scopeBucket -> scope -> acctBucket
	// scopeBucket -> scope -> addrBucket
	// scopeBucket -> scope -> usedAddrBucket
	// scopeBucket -> scope -> addrAcctIdxBucket
	// scopeBucket -> scope -> acctNameIdxBucket
	// scopeBucket -> scope -> acctIDIdxBucketName
	// scopeBucket -> scope -> metaBucket
	// scopeBucket -> scope -> metaBucket -> lastAccountNameKey
	// scopeBucket -> scope -> coinTypePrivKey
	// scopeBucket -> scope -> coinTypePubKey
	scopeBucketName = []byte("scope")

	// coinTypePrivKeyName is the name of the key within a particular scope
	// bucket that stores the encrypted cointype private keys. Each scope
	// within the database will have its own set of coin type keys.
	coinTypePrivKeyName = []byte("ctpriv")

	// coinTypePrivKeyName is the name of the key within a particular scope
	// bucket that stores the encrypted cointype public keys. Each scope
	// will have its own set of coin type public keys.
	coinTypePubKeyName = []byte("ctpub")

	// acctBucketName is the bucket directly below the scope bucket in the
	// hierarchy. This bucket stores all the information and indexes
	// relevant to an account.
	acctBucketName = []byte("acct")

	// addrBucketName is the name of the bucket that stores a mapping of
	// pubkey hash to address type. This will be used to quickly determine
	// if a given address is under our control.
	addrBucketName = []byte("addr")

	// addrAcctIdxBucketName is used to index account addresses Entries in
	// this index may map:
	// * addr hash => account id
	// * account bucket -> addr hash => null
	//
	// To fetch the account of an address, lookup the value using the
	// address hash.
	//
	// To fetch all addresses of an account, fetch the account bucket,
	// iterate over the keys and fetch the address row from the addr
	// bucket.
	//
	// The index needs to be updated whenever an address is created e.g.
	// NewAddress
	addrAcctIdxBucketName = []byte("addracctidx")

	// acctNameIdxBucketName is used to create an index mapping an account
	// name string to the corresponding account id.  The index needs to be
	// updated whenever the account name and id changes e.g. RenameAccount
	//
	// string => account_id
	acctNameIdxBucketName = []byte("acctnameidx")

	// acctIDIdxBucketName is used to create an index mapping an account id
	// to the corresponding account name string.  The index needs to be
	// updated whenever the account name and id changes e.g. RenameAccount
	//
	// account_id => string
	acctIDIdxBucketName = []byte("acctididx")

	// usedAddrBucketName is the name of the bucket that stores an
	// addresses hash if the address has been used or not.
	usedAddrBucketName = []byte("usedaddrs")

	// meta is used to store meta-data about the address manager
	// e.g. last account number
	metaBucketName = []byte("meta")

	// lastAccountName is used to store the metadata - last account
	// in the manager
	lastAccountName = []byte("lastaccount")

	// mainBucketName is the name of the bucket that stores the encrypted
	// crypto keys that encrypt all other generated keys, the watch only
	// flag, the master private key (encrypted), the master HD private key
	// (encrypted), and also versioning information.
	mainBucketName = []byte("main")

	// masterHDPrivName is the name of the key that stores the master HD
	// private key. This key is encrypted with the master private crypto
	// encryption key. This resides under the main bucket.
	masterHDPrivName = []byte("mhdpriv")

	// masterHDPubName is the name of the key that stores the master HD
	// public key. This key is encrypted with the master public crypto
	// encryption key. This reside under the main bucket.
	masterHDPubName = []byte("mhdpub")

	// syncBucketName is the name of the bucket that stores the current
	// sync state of the root manager.
	syncBucketName = []byte("sync")

	// Db related key names (main bucket).
	mgrVersionName    = []byte("mgrver")
	mgrCreateDateName = []byte("mgrcreated")

	// Crypto related key names (main bucket).
	masterPrivKeyName   = []byte("mpriv")
	masterPubKeyName    = []byte("mpub")
	cryptoPrivKeyName   = []byte("cpriv")
	cryptoPubKeyName    = []byte("cpub")
	cryptoScriptKeyName = []byte("cscript")
	watchingOnlyName    = []byte("watchonly")

	// Sync related key names (sync bucket).
	syncedToName              = []byte("syncedto")
	startBlockName            = []byte("startblock")
	birthdayName              = []byte("birthday")
	birthdayBlockName         = []byte("birthdayblock")
	birthdayBlockVerifiedName = []byte("birthdayblockverified")
)

// uint32ToBytes converts a 32 bit unsigned integer into a 4-byte slice in
// little-endian order: 1 -> [1 0 0 0].
func uint32ToBytes(number uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, number)
	return buf
}

// uint64ToBytes converts a 64 bit unsigned integer into a 8-byte slice in
// little-endian order: 1 -> [1 0 0 0 0 0 0 0].
func uint64ToBytes(number uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, number)
	return buf
}

// stringToBytes converts a string into a variable length byte slice in
// little-endian order: "abc" -> [3 0 0 0 61 62 63]
func stringToBytes(s string) []byte {
	// The serialized format is:
	//   <size><string>
	//
	// 4 bytes string size + string
	size := len(s)
	buf := make([]byte, 4+size)
	copy(buf[0:4], uint32ToBytes(uint32(size)))
	copy(buf[4:4+size], s)
	return buf
}

// scopeKeySize is the size of a scope as stored within the database.
const scopeKeySize = 8

// fetchManagerVersion fetches the current manager version from the database.
func fetchManagerVersion(ns walletdb.ReadBucket) (uint32, error) {
	mainBucket := ns.NestedReadBucket(mainBucketName)
	verBytes := mainBucket.Get(mgrVersionName)
	if verBytes == nil {
		str := "required version number not stored in database"
		return 0, managerError(ErrDatabase, str, nil)
	}
	version := binary.LittleEndian.Uint32(verBytes)
	return version, nil
}

// putManagerVersion stores the provided version to the database.
func putManagerVersion(ns walletdb.ReadWriteBucket, version uint32) error {
	bucket := ns.NestedReadWriteBucket(mainBucketName)

	verBytes := uint32ToBytes(version)
	err := bucket.Put(mgrVersionName, verBytes)
	if err != nil {
		str := "failed to store version"
		return managerError(ErrDatabase, str, err)
	}
	return nil
}

// fetchMasterKeyParams loads the master key parameters needed to derive them
// (when given the correct user-supplied passphrase) from the database.  Either
// returned value can be nil, but in practice only the private key params will
// be nil for a watching-only database.
func fetchMasterKeyParams(ns walletdb.ReadBucket) ([]byte, []byte, error) {
	bucket := ns.NestedReadBucket(mainBucketName)

	// Load the master public key parameters.  Required.
	val := bucket.Get(masterPubKeyName)
	if val == nil {
		str := "required master public key parameters not stored in " +
			"database"
		return nil, nil, managerError(ErrDatabase, str, nil)
	}
	pubParams := make([]byte, len(val))
	copy(pubParams, val)

	// Load the master private key parameters if they were stored.
	var privParams []byte
	val = bucket.Get(masterPrivKeyName)
	if val != nil {
		privParams = make([]byte, len(val))
		copy(privParams, val)
	}

	return pubParams, privParams, nil
}

// putMasterKeyParams stores the master key parameters needed to derive them to
// the database.  Either parameter can be nil in which case no value is
// written for the parameter.
func putMasterKeyParams(ns walletdb.ReadWriteBucket, pubParams, privParams []byte) error {
	bucket := ns.NestedReadWriteBucket(mainBucketName)

	if privParams != nil {
		err := bucket.Put(masterPrivKeyName, privParams)
		if err != nil {
			str := "failed to store master private key parameters"
			return managerError(ErrDatabase, str, err)
		}
	}

	if pubParams != nil {
		err := bucket.Put(masterPubKeyName, pubParams)
		if err != nil {
			str := "failed to store master public key parameters"
			return managerError(ErrDatabase, str, err)
		}
	}

	return nil
}

// putMasterHDKeys stores the encrypted master HD keys in the top level main
// bucket. These are required in order to create any new manager scopes, as
// those are created via hardened derivation of the children of this key.
func putMasterHDKeys(ns walletdb.ReadWriteBucket, masterHDPrivEnc, masterHDPubEnc []byte) error {
	// As this is the key for the root manager, we don't need to fetch any
	// particular scope, and can insert directly within the main bucket.
	bucket := ns.NestedReadWriteBucket(mainBucketName)

	// Now that we have the main bucket, we can directly store each of the
	// relevant keys. If we're in watch only mode, then some or all of
	// these keys might not be available.
	if masterHDPrivEnc != nil {
		err := bucket.Put(masterHDPrivName, masterHDPrivEnc)
		if err != nil {
			str := "failed to store encrypted master HD private key"
			return managerError(ErrDatabase, str, err)
		}
	}

	if masterHDPubEnc != nil {
		err := bucket.Put(masterHDPubName, masterHDPubEnc)
		if err != nil {
			str := "failed to store encrypted master HD public key"
			return managerError(ErrDatabase, str, err)
		}
	}

	return nil
}

// fetchMasterHDKeys attempts to fetch both the master HD private and public
// keys from the database. If this is a watch only wallet, then it's possible
// that the master private key isn't stored.
func fetchMasterHDKeys(ns walletdb.ReadBucket) ([]byte, []byte, error) {
	bucket := ns.NestedReadBucket(mainBucketName)

	var masterHDPrivEnc, masterHDPubEnc []byte

	// First, we'll try to fetch the master private key. If this database
	// is watch only, or the master has been neutered, then this won't be
	// found on disk.
	key := bucket.Get(masterHDPrivName)
	if key != nil {
		masterHDPrivEnc = make([]byte, len(key))
		copy(masterHDPrivEnc[:], key)
	}

	key = bucket.Get(masterHDPubName)
	if key != nil {
		masterHDPubEnc = make([]byte, len(key))
		copy(masterHDPubEnc[:], key)
	}

	return masterHDPrivEnc, masterHDPubEnc, nil
}

// fetchCryptoKeys loads the encrypted crypto keys which are in turn used to
// protect the extended keys, imported keys, and scripts.  Any of the returned
// values can be nil, but in practice only the crypto private and script keys
// will be nil for a watching-only database.
func fetchCryptoKeys(ns walletdb.ReadBucket) ([]byte, []byte, []byte, error) {
	bucket := ns.NestedReadBucket(mainBucketName)

	// Load the crypto public key parameters.  Required.
	val := bucket.Get(cryptoPubKeyName)
	if val == nil {
		str := "required encrypted crypto public not stored in database"
		return nil, nil, nil, managerError(ErrDatabase, str, nil)
	}
	pubKey := make([]byte, len(val))
	copy(pubKey, val)

	// Load the crypto private key parameters if they were stored.
	var privKey []byte
	val = bucket.Get(cryptoPrivKeyName)
	if val != nil {
		privKey = make([]byte, len(val))
		copy(privKey, val)
	}

	// Load the crypto script key parameters if they were stored.
	var scriptKey []byte
	val = bucket.Get(cryptoScriptKeyName)
	if val != nil {
		scriptKey = make([]byte, len(val))
		copy(scriptKey, val)
	}

	return pubKey, privKey, scriptKey, nil
}

// putCryptoKeys stores the encrypted crypto keys which are in turn used to
// protect the extended and imported keys.  Either parameter can be nil in
// which case no value is written for the parameter.
func putCryptoKeys(ns walletdb.ReadWriteBucket, pubKeyEncrypted, privKeyEncrypted,
	scriptKeyEncrypted []byte) error {

	bucket := ns.NestedReadWriteBucket(mainBucketName)

	if pubKeyEncrypted != nil {
		err := bucket.Put(cryptoPubKeyName, pubKeyEncrypted)
		if err != nil {
			str := "failed to store encrypted crypto public key"
			return managerError(ErrDatabase, str, err)
		}
	}

	if privKeyEncrypted != nil {
		err := bucket.Put(cryptoPrivKeyName, privKeyEncrypted)
		if err != nil {
			str := "failed to store encrypted crypto private key"
			return managerError(ErrDatabase, str, err)
		}
	}

	if scriptKeyEncrypted != nil {
		err := bucket.Put(cryptoScriptKeyName, scriptKeyEncrypted)
		if err != nil {
			str := "failed to store encrypted crypto script key"
			return managerError(ErrDatabase, str, err)
		}
	}

	return nil
}

// fetchWatchingOnly loads the watching-only flag from the database.
func fetchWatchingOnly(ns walletdb.ReadBucket) (bool, error) {
	bucket := ns.NestedReadBucket(mainBucketName)

	buf := bucket.Get(watchingOnlyName)
	if len(buf) != 1 {
		str := "malformed watching-only flag stored in database"
		return false, managerError(ErrDatabase, str, nil)
	}

	return buf[0] != 0, nil
}

// putWatchingOnly stores the watching-only flag to the database.
func putWatchingOnly(ns walletdb.ReadWriteBucket, watchingOnly bool) error {
	bucket := ns.NestedReadWriteBucket(mainBucketName)

	var encoded byte
	if watchingOnly {
		encoded = 1
	}

	if err := bucket.Put(watchingOnlyName, []byte{encoded}); err != nil {
		str := "failed to store watching only flag"
		return managerError(ErrDatabase, str, err)
	}
	return nil
}

// deserializeAccountRow deserializes the passed serialized account information.
// This is used as a common base for the various account types to deserialize
// the common parts.
func deserializeAccountRow(accountID []byte, serializedAccount []byte) (*dbAccountRow, error) {
	// The serialized account format is:
	//   <acctType><rdlen><rawdata>
	//
	// 1 byte acctType + 4 bytes raw data length + raw data

	// Given the above, the length of the entry must be at a minimum
	// the constant value sizes.
	if len(serializedAccount) < 5 {
		str := fmt.Sprintf("malformed serialized account for key %x",
			accountID)
		return nil, managerError(ErrDatabase, str, nil)
	}

	row := dbAccountRow{}
	row.acctType = accountType(serializedAccount[0])
	rdlen := binary.LittleEndian.Uint32(serializedAccount[1:5])
	row.rawData = make([]byte, rdlen)
	copy(row.rawData, serializedAccount[5:5+rdlen])

	return &row, nil
}

// serializeAccountRow returns the serialization of the passed account row.
func serializeAccountRow(row *dbAccountRow) []byte {
	// The serialized account format is:
	//   <acctType><rdlen><rawdata>
	//
	// 1 byte acctType + 4 bytes raw data length + raw data
	rdlen := len(row.rawData)
	buf := make([]byte, 5+rdlen)
	buf[0] = byte(row.acctType)
	binary.LittleEndian.PutUint32(buf[1:5], uint32(rdlen))
	copy(buf[5:5+rdlen], row.rawData)
	return buf
}

// deserializeDefaultAccountRow deserializes the raw data from the passed
// account row as a BIP0044-like account.
func deserializeDefaultAccountRow(accountID []byte, row *dbAccountRow) (*dbDefaultAccountRow, error) {
	// The serialized BIP0044 account raw data format is:
	//   <encpubkeylen><encpubkey><encprivkeylen><encprivkey><nextextidx>
	//   <nextintidx><namelen><name>
	//
	// 4 bytes encrypted pubkey len + encrypted pubkey + 4 bytes encrypted
	// privkey len + encrypted privkey + 4 bytes next external index +
	// 4 bytes next internal index + 4 bytes name len + name

	// Given the above, the length of the entry must be at a minimum
	// the constant value sizes.
	if len(row.rawData) < 20 {
		str := fmt.Sprintf("malformed serialized bip0044 account for "+
			"key %x", accountID)
		return nil, managerError(ErrDatabase, str, nil)
	}

	retRow := dbDefaultAccountRow{
		dbAccountRow: *row,
	}

	pubLen := binary.LittleEndian.Uint32(row.rawData[0:4])
	retRow.pubKeyEncrypted = make([]byte, pubLen)
	copy(retRow.pubKeyEncrypted, row.rawData[4:4+pubLen])
	offset := 4 + pubLen
	privLen := binary.LittleEndian.Uint32(row.rawData[offset : offset+4])
	offset += 4
	retRow.privKeyEncrypted = make([]byte, privLen)
	copy(retRow.privKeyEncrypted, row.rawData[offset:offset+privLen])
	offset += privLen
	retRow.nextExternalIndex = binary.LittleEndian.Uint32(row.rawData[offset : offset+4])
	offset += 4
	retRow.nextInternalIndex = binary.LittleEndian.Uint32(row.rawData[offset : offset+4])
	offset += 4
	nameLen := binary.LittleEndian.Uint32(row.rawData[offset : offset+4])
	offset += 4
	retRow.name = string(row.rawData[offset : offset+nameLen])

	return &retRow, nil
}

// serializeDefaultAccountRow returns the serialization of the raw data field
// for a BIP0044-like account.
func serializeDefaultAccountRow(encryptedPubKey, encryptedPrivKey []byte,
	nextExternalIndex, nextInternalIndex uint32, name string) []byte {

	// The serialized BIP0044 account raw data format is:
	//   <encpubkeylen><encpubkey><encprivkeylen><encprivkey><nextextidx>
	//   <nextintidx><namelen><name>
	//
	// 4 bytes encrypted pubkey len + encrypted pubkey + 4 bytes encrypted
	// privkey len + encrypted privkey + 4 bytes next external index +
	// 4 bytes next internal index + 4 bytes name len + name
	pubLen := uint32(len(encryptedPubKey))
	privLen := uint32(len(encryptedPrivKey))
	nameLen := uint32(len(name))
	rawData := make([]byte, 20+pubLen+privLen+nameLen)
	binary.LittleEndian.PutUint32(rawData[0:4], pubLen)
	copy(rawData[4:4+pubLen], encryptedPubKey)
	offset := 4 + pubLen
	binary.LittleEndian.PutUint32(rawData[offset:offset+4], privLen)
	offset += 4
	copy(rawData[offset:offset+privLen], encryptedPrivKey)
	offset += privLen
	binary.LittleEndian.PutUint32(rawData[offset:offset+4], nextExternalIndex)
	offset += 4
	binary.LittleEndian.PutUint32(rawData[offset:offset+4], nextInternalIndex)
	offset += 4
	binary.LittleEndian.PutUint32(rawData[offset:offset+4], nameLen)
	offset += 4
	copy(rawData[offset:offset+nameLen], name)
	return rawData
}


// deserializeAddressRow deserializes the passed serialized address
// information.  This is used as a common base for the various address types to
// deserialize the common parts.
func deserializeAddressRow(serializedAddress []byte) (*dbAddressRow, error) {
	// The serialized address format is:
	//   <addrType><account><addedTime><syncStatus><rawdata>
	//
	// 1 byte addrType + 4 bytes account + 8 bytes addTime + 1 byte
	// syncStatus + 4 bytes raw data length + raw data

	// Given the above, the length of the entry must be at a minimum
	// the constant value sizes.
	if len(serializedAddress) < 18 {
		str := "malformed serialized address"
		return nil, managerError(ErrDatabase, str, nil)
	}

	row := dbAddressRow{}
	row.addrType = addressType(serializedAddress[0])
	row.account = binary.LittleEndian.Uint32(serializedAddress[1:5])
	row.addTime = binary.LittleEndian.Uint64(serializedAddress[5:13])
	row.syncStatus = syncStatus(serializedAddress[13])
	rdlen := binary.LittleEndian.Uint32(serializedAddress[14:18])
	row.rawData = make([]byte, rdlen)
	copy(row.rawData, serializedAddress[18:18+rdlen])

	return &row, nil
}

// serializeAddressRow returns the serialization of the passed address row.
func serializeAddressRow(row *dbAddressRow) []byte {
	// The serialized address format is:
	//   <addrType><account><addedTime><syncStatus><commentlen><comment>
	//   <rawdata>
	//
	// 1 byte addrType + 4 bytes account + 8 bytes addTime + 1 byte
	// syncStatus + 4 bytes raw data length + raw data
	rdlen := len(row.rawData)
	buf := make([]byte, 18+rdlen)
	buf[0] = byte(row.addrType)
	binary.LittleEndian.PutUint32(buf[1:5], row.account)
	binary.LittleEndian.PutUint64(buf[5:13], row.addTime)
	buf[13] = byte(row.syncStatus)
	binary.LittleEndian.PutUint32(buf[14:18], uint32(rdlen))
	copy(buf[18:18+rdlen], row.rawData)
	return buf
}

// deserializeChainedAddress deserializes the raw data from the passed address
// row as a chained address.
func deserializeChainedAddress(row *dbAddressRow) (*dbChainAddressRow, error) {
	// The serialized chain address raw data format is:
	//   <branch><index>
	//
	// 4 bytes branch + 4 bytes address index
	if len(row.rawData) != 8 {
		str := "malformed serialized chained address"
		return nil, managerError(ErrDatabase, str, nil)
	}

	retRow := dbChainAddressRow{
		dbAddressRow: *row,
	}

	retRow.branch = binary.LittleEndian.Uint32(row.rawData[0:4])
	retRow.index = binary.LittleEndian.Uint32(row.rawData[4:8])

	return &retRow, nil
}

// serializeChainedAddress returns the serialization of the raw data field for
// a chained address.
func serializeChainedAddress(branch, index uint32) []byte {
	// The serialized chain address raw data format is:
	//   <branch><index>
	//
	// 4 bytes branch + 4 bytes address index
	rawData := make([]byte, 8)
	binary.LittleEndian.PutUint32(rawData[0:4], branch)
	binary.LittleEndian.PutUint32(rawData[4:8], index)
	return rawData
}

// deserializeImportedAddress deserializes the raw data from the passed address
// row as an imported address.
func deserializeImportedAddress(row *dbAddressRow) (*dbImportedAddressRow, error) {
	// The serialized imported address raw data format is:
	//   <encpubkeylen><encpubkey><encprivkeylen><encprivkey>
	//
	// 4 bytes encrypted pubkey len + encrypted pubkey + 4 bytes encrypted
	// privkey len + encrypted privkey

	// Given the above, the length of the entry must be at a minimum
	// the constant value sizes.
	if len(row.rawData) < 8 {
		str := "malformed serialized imported address"
		return nil, managerError(ErrDatabase, str, nil)
	}

	retRow := dbImportedAddressRow{
		dbAddressRow: *row,
	}

	pubLen := binary.LittleEndian.Uint32(row.rawData[0:4])
	retRow.encryptedPubKey = make([]byte, pubLen)
	copy(retRow.encryptedPubKey, row.rawData[4:4+pubLen])
	offset := 4 + pubLen
	privLen := binary.LittleEndian.Uint32(row.rawData[offset : offset+4])
	offset += 4
	retRow.encryptedPrivKey = make([]byte, privLen)
	copy(retRow.encryptedPrivKey, row.rawData[offset:offset+privLen])

	return &retRow, nil
}

// serializeImportedAddress returns the serialization of the raw data field for
// an imported address.
func serializeImportedAddress(encryptedPubKey, encryptedPrivKey []byte) []byte {
	// The serialized imported address raw data format is:
	//   <encpubkeylen><encpubkey><encprivkeylen><encprivkey>
	//
	// 4 bytes encrypted pubkey len + encrypted pubkey + 4 bytes encrypted
	// privkey len + encrypted privkey
	pubLen := uint32(len(encryptedPubKey))
	privLen := uint32(len(encryptedPrivKey))
	rawData := make([]byte, 8+pubLen+privLen)
	binary.LittleEndian.PutUint32(rawData[0:4], pubLen)
	copy(rawData[4:4+pubLen], encryptedPubKey)
	offset := 4 + pubLen
	binary.LittleEndian.PutUint32(rawData[offset:offset+4], privLen)
	offset += 4
	copy(rawData[offset:offset+privLen], encryptedPrivKey)
	return rawData
}

// deserializeScriptAddress deserializes the raw data from the passed address
// row as a script address.
func deserializeScriptAddress(row *dbAddressRow) (*dbScriptAddressRow, error) {
	// The serialized script address raw data format is:
	//   <encscripthashlen><encscripthash><encscriptlen><encscript>
	//
	// 4 bytes encrypted script hash len + encrypted script hash + 4 bytes
	// encrypted script len + encrypted script

	// Given the above, the length of the entry must be at a minimum
	// the constant value sizes.
	if len(row.rawData) < 8 {
		str := "malformed serialized script address"
		return nil, managerError(ErrDatabase, str, nil)
	}

	retRow := dbScriptAddressRow{
		dbAddressRow: *row,
	}

	hashLen := binary.LittleEndian.Uint32(row.rawData[0:4])
	retRow.encryptedHash = make([]byte, hashLen)
	copy(retRow.encryptedHash, row.rawData[4:4+hashLen])
	offset := 4 + hashLen
	scriptLen := binary.LittleEndian.Uint32(row.rawData[offset : offset+4])
	offset += 4
	retRow.encryptedScript = make([]byte, scriptLen)
	copy(retRow.encryptedScript, row.rawData[offset:offset+scriptLen])

	return &retRow, nil
}

// serializeScriptAddress returns the serialization of the raw data field for
// a script address.
func serializeScriptAddress(encryptedHash, encryptedScript []byte) []byte {
	// The serialized script address raw data format is:
	//   <encscripthashlen><encscripthash><encscriptlen><encscript>
	//
	// 4 bytes encrypted script hash len + encrypted script hash + 4 bytes
	// encrypted script len + encrypted script

	hashLen := uint32(len(encryptedHash))
	scriptLen := uint32(len(encryptedScript))
	rawData := make([]byte, 8+hashLen+scriptLen)
	binary.LittleEndian.PutUint32(rawData[0:4], hashLen)
	copy(rawData[4:4+hashLen], encryptedHash)
	offset := 4 + hashLen
	binary.LittleEndian.PutUint32(rawData[offset:offset+4], scriptLen)
	offset += 4
	copy(rawData[offset:offset+scriptLen], encryptedScript)
	return rawData
}

// deletePrivateKeys removes all private key material from the database.
//
// NOTE: Care should be taken when calling this function.  It is primarily
// intended for use in converting to a watching-only copy.  Removing the private
// keys from the main database without also marking it watching-only will result
// in an unusable database.  It will also make any imported scripts and private
// keys unrecoverable unless there is a backup copy available.
func deletePrivateKeys(ns walletdb.ReadWriteBucket) error {
	bucket := ns.NestedReadWriteBucket(mainBucketName)

	// Delete the master private key params and the crypto private and
	// script keys.
	if err := bucket.Delete(masterPrivKeyName); err != nil {
		str := "failed to delete master private key parameters"
		return managerError(ErrDatabase, str, err)
	}
	if err := bucket.Delete(cryptoPrivKeyName); err != nil {
		str := "failed to delete crypto private key"
		return managerError(ErrDatabase, str, err)
	}
	if err := bucket.Delete(cryptoScriptKeyName); err != nil {
		str := "failed to delete crypto script key"
		return managerError(ErrDatabase, str, err)
	}
	if err := bucket.Delete(masterHDPrivName); err != nil {
		str := "failed to delete master HD priv key"
		return managerError(ErrDatabase, str, err)
	}

	// With the master key and meta encryption keys deleted, we'll need to
	// delete the keys for all known scopes as well.
	scopeBucket := ns.NestedReadWriteBucket(scopeBucketName)
	err := scopeBucket.ForEach(func(scopeKey, _ []byte) error {
		if len(scopeKey) != 8 {
			return nil
		}

		managerScopeBucket := scopeBucket.NestedReadWriteBucket(scopeKey)

		if err := managerScopeBucket.Delete(coinTypePrivKeyName); err != nil {
			str := "failed to delete cointype private key"
			return managerError(ErrDatabase, str, err)
		}

		// Delete the account extended private key for all accounts.
		bucket = managerScopeBucket.NestedReadWriteBucket(acctBucketName)
		err := bucket.ForEach(func(k, v []byte) error {
			// Skip buckets.
			if v == nil {
				return nil
			}

			// Deserialize the account row first to determine the type.
			row, err := deserializeAccountRow(k, v)
			if err != nil {
				return err
			}

			switch row.acctType {
			case accountDefault:
				arow, err := deserializeDefaultAccountRow(k, row)
				if err != nil {
					return err
				}

				// Reserialize the account without the private key and
				// store it.
				row.rawData = serializeDefaultAccountRow(
					arow.pubKeyEncrypted, nil,
					arow.nextExternalIndex, arow.nextInternalIndex,
					arow.name,
				)
				err = bucket.Put(k, serializeAccountRow(row))
				if err != nil {
					str := "failed to delete account private key"
					return managerError(ErrDatabase, str, err)
				}
			}

			return nil
		})
		if err != nil {
			return maybeConvertDbError(err)
		}

		// Delete the private key for all imported addresses.
		bucket = managerScopeBucket.NestedReadWriteBucket(addrBucketName)
		err = bucket.ForEach(func(k, v []byte) error {
			// Skip buckets.
			if v == nil {
				return nil
			}

			// Deserialize the address row first to determine the field
			// values.
			row, err := deserializeAddressRow(v)
			if err != nil {
				return err
			}

			switch row.addrType {
			case adtImport:
				irow, err := deserializeImportedAddress(row)
				if err != nil {
					return err
				}

				// Reserialize the imported address without the private
				// key and store it.
				row.rawData = serializeImportedAddress(
					irow.encryptedPubKey, nil)
				err = bucket.Put(k, serializeAddressRow(row))
				if err != nil {
					str := "failed to delete imported private key"
					return managerError(ErrDatabase, str, err)
				}

			case adtScript:
				srow, err := deserializeScriptAddress(row)
				if err != nil {
					return err
				}

				// Reserialize the script address without the script
				// and store it.
				row.rawData = serializeScriptAddress(srow.encryptedHash,
					nil)
				err = bucket.Put(k, serializeAddressRow(row))
				if err != nil {
					str := "failed to delete imported script"
					return managerError(ErrDatabase, str, err)
				}
			}

			return nil
		})
		if err != nil {
			return maybeConvertDbError(err)
		}

		return nil
	})
	if err != nil {
		return maybeConvertDbError(err)
	}

	return nil
}

// fetchSyncedTo loads the block stamp the manager is synced to from the
// database.
func fetchSyncedTo(ns walletdb.ReadBucket) (*BlockStamp, error) {
	bucket := ns.NestedReadBucket(syncBucketName)

	// The serialized synced to format is:
	//   <blockheight><blockhash><timestamp>
	//
	// 4 bytes block height + 32 bytes hash length
	buf := bucket.Get(syncedToName)
	if len(buf) < 36 {
		str := "malformed sync information stored in database"
		return nil, managerError(ErrDatabase, str, nil)
	}

	var bs BlockStamp
	bs.Height = int32(binary.LittleEndian.Uint32(buf[0:4]))
	copy(bs.Hash[:], buf[4:36])

	if len(buf) == 40 {
		bs.Timestamp = time.Unix(
			int64(binary.LittleEndian.Uint32(buf[36:])), 0,
		)
	}

	return &bs, nil
}

// PutSyncedTo stores the provided synced to blockstamp to the database.
func PutSyncedTo(ns walletdb.ReadWriteBucket, bs *BlockStamp) error {
	bucket := ns.NestedReadWriteBucket(syncBucketName)
	errStr := fmt.Sprintf("failed to store sync information %v", bs.Hash)

	// If the block height is greater than zero, check that the previous
	// block height exists. This prevents reorg issues in the future.
	// We use BigEndian so that keys/values are added to the bucket in
	// order, making writes more efficient for some database backends.
	if bs.Height > 0 {
		if _, err := fetchBlockHash(ns, bs.Height-1); err != nil {
			return managerError(ErrDatabase, errStr, err)
		}
	}

	// Store the block hash by block height.
	height := make([]byte, 4)
	binary.BigEndian.PutUint32(height, uint32(bs.Height))
	err := bucket.Put(height, bs.Hash[0:32])
	if err != nil {
		return managerError(ErrDatabase, errStr, err)
	}

	// The serialized synced to format is:
	//   <blockheight><blockhash><timestamp>
	//
	// 4 bytes block height + 32 bytes hash length + 4 byte timestamp length
	buf := make([]byte, 40)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(bs.Height))
	copy(buf[4:36], bs.Hash[0:32])
	binary.LittleEndian.PutUint32(buf[36:], uint32(bs.Timestamp.Unix()))

	err = bucket.Put(syncedToName, buf)
	if err != nil {
		return managerError(ErrDatabase, errStr, err)
	}
	return nil
}

// fetchBlockHash loads the block hash for the provided height from the
// database.
func fetchBlockHash(ns walletdb.ReadBucket, height int32) (*chainhash.Hash, error) {
	bucket := ns.NestedReadBucket(syncBucketName)
	errStr := fmt.Sprintf("failed to fetch block hash for height %d", height)

	heightBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(heightBytes, uint32(height))
	hashBytes := bucket.Get(heightBytes)
	if hashBytes == nil {
		err := errors.New("block not found")
		return nil, managerError(ErrBlockNotFound, errStr, err)
	}
	if len(hashBytes) != 32 {
		err := fmt.Errorf("couldn't get hash from database")
		return nil, managerError(ErrDatabase, errStr, err)
	}
	var hash chainhash.Hash
	if err := hash.SetBytes(hashBytes); err != nil {
		return nil, managerError(ErrDatabase, errStr, err)
	}
	return &hash, nil
}

// FetchStartBlock loads the start block stamp for the manager from the
// database.
func FetchStartBlock(ns walletdb.ReadBucket) (*BlockStamp, error) {
	bucket := ns.NestedReadBucket(syncBucketName)

	// The serialized start block format is:
	//   <blockheight><blockhash>
	//
	// 4 bytes block height + 32 bytes hash length
	buf := bucket.Get(startBlockName)
	if len(buf) != 36 {
		str := "malformed start block stored in database"
		return nil, managerError(ErrDatabase, str, nil)
	}

	var bs BlockStamp
	bs.Height = int32(binary.LittleEndian.Uint32(buf[0:4]))
	copy(bs.Hash[:], buf[4:36])
	return &bs, nil
}

// putStartBlock stores the provided start block stamp to the database.
func putStartBlock(ns walletdb.ReadWriteBucket, bs *BlockStamp) error {
	bucket := ns.NestedReadWriteBucket(syncBucketName)

	// The serialized start block format is:
	//   <blockheight><blockhash>
	//
	// 4 bytes block height + 32 bytes hash length
	buf := make([]byte, 36)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(bs.Height))
	copy(buf[4:36], bs.Hash[0:32])

	err := bucket.Put(startBlockName, buf)
	if err != nil {
		str := fmt.Sprintf("failed to store start block %v", bs.Hash)
		return managerError(ErrDatabase, str, err)
	}
	return nil
}

// fetchBirthday loads the manager's bithday timestamp from the database.
func fetchBirthday(ns walletdb.ReadBucket) (time.Time, error) {
	var t time.Time

	bucket := ns.NestedReadBucket(syncBucketName)
	birthdayTimestamp := bucket.Get(birthdayName)
	if len(birthdayTimestamp) != 8 {
		str := "malformed birthday stored in database"
		return t, managerError(ErrDatabase, str, nil)
	}

	t = time.Unix(int64(binary.BigEndian.Uint64(birthdayTimestamp)), 0)

	return t, nil
}

// putBirthday stores the provided birthday timestamp to the database.
func putBirthday(ns walletdb.ReadWriteBucket, t time.Time) error {
	var birthdayTimestamp [8]byte
	binary.BigEndian.PutUint64(birthdayTimestamp[:], uint64(t.Unix()))

	bucket := ns.NestedReadWriteBucket(syncBucketName)
	if err := bucket.Put(birthdayName, birthdayTimestamp[:]); err != nil {
		str := "failed to store birthday"
		return managerError(ErrDatabase, str, err)
	}

	return nil
}

// FetchBirthdayBlock retrieves the birthday block from the database.
//
// The block is serialized as follows:
//   [0:4]   block height
//   [4:36]  block hash
//   [36:44] block timestamp
func FetchBirthdayBlock(ns walletdb.ReadBucket) (BlockStamp, error) {
	var block BlockStamp

	bucket := ns.NestedReadBucket(syncBucketName)
	birthdayBlock := bucket.Get(birthdayBlockName)
	if birthdayBlock == nil {
		str := "birthday block not set"
		return block, managerError(ErrBirthdayBlockNotSet, str, nil)
	}
	if len(birthdayBlock) != 44 {
		str := "malformed birthday block stored in database"
		return block, managerError(ErrDatabase, str, nil)
	}

	block.Height = int32(binary.BigEndian.Uint32(birthdayBlock[:4]))
	copy(block.Hash[:], birthdayBlock[4:36])
	t := int64(binary.BigEndian.Uint64(birthdayBlock[36:]))
	block.Timestamp = time.Unix(t, 0)

	return block, nil
}

// putBirthdayBlock stores the provided birthday block to the database.
//
// The block is serialized as follows:
//   [0:4]   block height
//   [4:36]  block hash
//   [36:44] block timestamp
func putBirthdayBlock(ns walletdb.ReadWriteBucket, block BlockStamp) error {
	var birthdayBlock [44]byte
	binary.BigEndian.PutUint32(birthdayBlock[:4], uint32(block.Height))
	copy(birthdayBlock[4:36], block.Hash[:])
	binary.BigEndian.PutUint64(birthdayBlock[36:], uint64(block.Timestamp.Unix()))

	bucket := ns.NestedReadWriteBucket(syncBucketName)
	if err := bucket.Put(birthdayBlockName, birthdayBlock[:]); err != nil {
		str := "failed to store birthday block"
		return managerError(ErrDatabase, str, err)
	}

	return nil
}

// fetchBirthdayBlockVerification retrieves the bit that determines whether the
// wallet has verified that its birthday block is correct.
func fetchBirthdayBlockVerification(ns walletdb.ReadBucket) bool {
	bucket := ns.NestedReadBucket(syncBucketName)
	verifiedValue := bucket.Get(birthdayBlockVerifiedName)

	// If there is no verification status, we can assume it has not been
	// verified yet.
	if verifiedValue == nil {
		return false
	}

	// Otherwise, we'll determine if it's verified by the value stored.
	verified := binary.BigEndian.Uint16(verifiedValue[:])
	return verified != 0
}

// putBirthdayBlockVerification stores a bit that determines whether the
// birthday block has been verified by the wallet to be correct.
func putBirthdayBlockVerification(ns walletdb.ReadWriteBucket, verified bool) error {
	// Convert the boolean to an integer in its binary representation as
	// there is no way to insert a boolean directly as a value of a
	// key/value pair.
	verifiedValue := uint16(0)
	if verified {
		verifiedValue = 1
	}

	var verifiedBytes [2]byte
	binary.BigEndian.PutUint16(verifiedBytes[:], verifiedValue)

	bucket := ns.NestedReadWriteBucket(syncBucketName)
	err := bucket.Put(birthdayBlockVerifiedName, verifiedBytes[:])
	if err != nil {
		str := "failed to store birthday block verification"
		return managerError(ErrDatabase, str, err)
	}

	return nil
}

// managerExists returns whether or not the manager has already been created
// in the given database namespace.
func managerExists(ns walletdb.ReadBucket) bool {
	if ns == nil {
		return false
	}
	mainBucket := ns.NestedReadBucket(mainBucketName)
	return mainBucket != nil
}
// createManagerNS creates the initial namespace structure needed for all of
// the manager data.  This includes things such as all of the buckets as well
// as the version and creation date. In addition to creating the key space for
// the root address manager, we'll also create internal scopes for all the
// default manager scope types.
func createManagerNS(ns walletdb.ReadWriteBucket) error {

	// First, we'll create all the relevant buckets that stem off of the
	// main bucket.
	mainBucket, err := ns.CreateBucket(mainBucketName)
	if err != nil {
		str := "failed to create main bucket"
		return managerError(ErrDatabase, str, err)
	}
	_, err = ns.CreateBucket(syncBucketName)
	if err != nil {
		str := "failed to create sync bucket"
		return managerError(ErrDatabase, str, err)
	}

	if err := putManagerVersion(ns, latestMgrVersion); err != nil {
		return err
	}

	createDate := uint64(time.Now().Unix())
	var dateBytes [8]byte
	binary.LittleEndian.PutUint64(dateBytes[:], createDate)
	err = mainBucket.Put(mgrCreateDateName, dateBytes[:])
	if err != nil {
		str := "failed to store database creation time"
		return managerError(ErrDatabase, str, err)
	}

	return nil
}
