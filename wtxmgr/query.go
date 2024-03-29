package wtxmgr

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/walletdb"
	"github.com/Qitmeer/qng/common/hash"
	"github.com/Qitmeer/qng/core/types"
)

// CreditRecord contains metadata regarding a transaction credit for a known
// transaction.  Further details may be looked up by indexing a wire.MsgTx.TxOut
// with the Index field.
type CreditRecord struct {
	Amount types.Amount
	Index  uint32
	Spent  bool
	Change bool
}

// DebitRecord contains metadata regarding a transaction debit for a known
// transaction.  Further details may be looked up by indexing a wire.MsgTx.TxIn
// with the Index field.
type DebitRecord struct {
	Amount types.Amount
	Index  uint32
}

// TxDetails is intended to provide callers with access to rich details
// regarding a relevant transaction and which inputs and outputs are credit or
// debits.
type TxDetails struct {
	TxRecord
	Block   BlockMeta
	Credits []CreditRecord
	Debits  []DebitRecord
}

// minedTxDetails fetches the TxDetails for the mined transaction with hash
// txHash and the passed tx record key and value.
func (s *Store) minedTxDetails(ns walletdb.ReadBucket, txHash *hash.Hash, recKey, recVal []byte) (*TxDetails, error) {
	var details TxDetails

	// Parse transaction record k/v, lookup the full block record for the
	// block time, and read all matching credits, debits.
	err := readRawTxRecord(txHash, recVal, &details.TxRecord)
	if err != nil {
		return nil, err
	}
	err = readRawTxRecordBlock(recKey, &details.Block.Block)
	if err != nil {
		return nil, err
	}
	details.Block.Time, err = fetchBlockTime(ns, uint32(details.Block.Order))
	if err != nil {
		return nil, err
	}

	credIter := makeReadCreditIterator(ns, recKey)
	for credIter.next() {
		if int(credIter.elem.Index) >= len(details.MsgTx.TxOut) {
			str := "saved credit index exceeds number of outputs"
			return nil, storeError(ErrData, str, nil)
		}

		// The credit iterator does not record whether this credit was
		// spent by an unMined transaction, so check that here.
		if !credIter.elem.Spent {
			k := canonicalOutPoint(txHash, credIter.elem.Index)
			spent := existsRawUnMinedInput(ns, k) != nil
			credIter.elem.Spent = spent
		}
		details.Credits = append(details.Credits, credIter.elem)
	}
	if credIter.err != nil {
		return nil, credIter.err
	}

	debIter := makeReadDebitIterator(ns, recKey)
	for debIter.next() {
		if int(debIter.elem.Index) >= len(details.MsgTx.TxIn) {
			str := "saved debit index exceeds number of inputs"
			return nil, storeError(ErrData, str, nil)
		}

		details.Debits = append(details.Debits, debIter.elem)
	}
	return &details, debIter.err
}

// unMinedTxDetails fetches the TxDetails for the unMined transaction with the
// hash txHash and the passed unMined record value.
func (s *Store) unMinedTxDetails(ns walletdb.ReadBucket, txHash *hash.Hash, v []byte) (*TxDetails, error) {
	details := TxDetails{
		Block: BlockMeta{Block: Block{Order: -1}},
	}
	err := readRawTxRecord(txHash, v, &details.TxRecord)
	if err != nil {
		return nil, err
	}

	it := makeReadUnMinedCreditIterator(ns, txHash)
	for it.next() {
		if int(it.elem.Index) >= len(details.MsgTx.TxOut) {
			str := "saved credit index exceeds number of outputs"
			return nil, storeError(ErrData, str, nil)
		}

		// Set the Spent field since this is not done by the iterator.
		it.elem.Spent = existsRawUnMinedInput(ns, it.ck) != nil
		details.Credits = append(details.Credits, it.elem)
	}
	if it.err != nil {
		return nil, it.err
	}

	for i, output := range details.MsgTx.TxIn {
		opKey := canonicalOutPoint(&output.PreviousOut.Hash,
			output.PreviousOut.OutIndex)
		credKey := existsRawUnspent(ns, opKey)
		if credKey != nil {
			v := existsRawCredit(ns, credKey)
			amount, err := fetchRawCreditAmount(v)
			if err != nil {
				return nil, err
			}

			details.Debits = append(details.Debits, DebitRecord{
				Amount: amount,
				Index:  uint32(i),
			})
			continue
		}

		v := existsRawUnMinedCredit(ns, opKey)
		if v == nil {
			continue
		}

		amount, err := fetchRawCreditAmount(v)
		if err != nil {
			return nil, err
		}
		details.Debits = append(details.Debits, DebitRecord{
			Amount: amount,
			Index:  uint32(i),
		})
	}

	return &details, nil
}

// TxDetails looks up all recorded details regarding a transaction with some
// hash.  In case of a hash collision, the most recent transaction with a
// matching hash is returned.
//
// Not finding a transaction with this hash is not an error.  In this case,
// a nil TxDetails is returned.
func (s *Store) TxDetails(ns walletdb.ReadBucket, txHash *hash.Hash) (*TxDetails, error) {
	// First, check whether there exists an unMined transaction with this
	// hash.  Use it if found.
	v := existsRawUnMined(ns, txHash[:])
	if v != nil {
		return s.unMinedTxDetails(ns, txHash, v)
	}

	// Otherwise, if there exists a mined transaction with this matching
	// hash, skip over to the newest and begin fetching all details.
	k, v := latestTxRecord(ns, txHash)
	if v == nil {
		// not found
		return nil, nil
	}
	return s.minedTxDetails(ns, txHash, k, v)
}

// UniqueTxDetails looks up all recorded details for a transaction recorded
// mined in some particular block, or an unMined transaction if block is nil.
//
// Not finding a transaction with this hash from this block is not an error.  In
// this case, a nil TxDetails is returned.
func (s *Store) UniqueTxDetails(ns walletdb.ReadBucket, txHash *hash.Hash,
	block *Block) (*TxDetails, error) {

	if block == nil {
		v := existsRawUnMined(ns, txHash[:])
		if v == nil {
			return nil, nil
		}
		return s.unMinedTxDetails(ns, txHash, v)
	}

	k, v := existsTxRecord(ns, txHash, block)
	if v == nil {
		return nil, nil
	}
	return s.minedTxDetails(ns, txHash, k, v)
}

// rangeUnMinedTransactions executes the function f with TxDetails for every
// unMined transaction.  f is not executed if no unMined transactions exist.
// Error returns from f (if any) are propigated to the caller.  Returns true
// (signaling breaking out of a RangeTransactions) iff f executes and returns
// true.
func (s *Store) rangeUnMinedTransactions(ns walletdb.ReadBucket, f func([]TxDetails) (bool, error)) (bool, error) {
	var details []TxDetails
	err := ns.NestedReadBucket(bucketUnMined).ForEach(func(k, v []byte) error {
		if len(k) < 32 {
			str := fmt.Sprintf("%s: short key (expected %d "+
				"bytes, read %d)", bucketUnMined, 32, len(k))
			return storeError(ErrData, str, nil)
		}

		var txHash hash.Hash
		copy(txHash[:], k)
		detail, err := s.unMinedTxDetails(ns, &txHash, v)
		if err != nil {
			return err
		}

		// Because the key was created while foreach-ing over the
		// bucket, it should be impossible for unMinedTxDetails to ever
		// successfully return a nil details struct.
		details = append(details, *detail)
		return nil
	})
	if err == nil && len(details) > 0 {
		return f(details)
	}
	return false, err
}

// rangeBlockTransactions executes the function f with TxDetails for every block
// between heights begin and end (reverse order when end > begin) until f
// returns true, or the transactions from block is processed.  Returns true iff
// f executes and returns true.
func (s *Store) rangeBlockTransactions(ns walletdb.ReadBucket, begin, end int32,
	f func([]TxDetails) (bool, error)) (bool, error) {

	if begin < 0 {
		begin = int32(^uint32(0) >> 1)
	}
	if end < 0 {
		end = int32(^uint32(0) >> 1)
	}

	var blockIter blockIterator
	var advance func(*blockIterator) bool
	if begin < end {
		// Iterate in forwards order
		blockIter = makeReadBlockIterator(ns, begin)
		advance = func(it *blockIterator) bool {
			if !it.next() {
				return false
			}
			return it.elem.Order <= end
		}
	} else {
		// Iterate in backwards order, from begin -> end.
		blockIter = makeReadBlockIterator(ns, begin)
		advance = func(it *blockIterator) bool {
			if !it.prev() {
				return false
			}
			return end <= it.elem.Order
		}
	}

	var details []TxDetails
	for advance(&blockIter) {
		block := &blockIter.elem

		if cap(details) < len(block.transactions) {
			details = make([]TxDetails, 0, len(block.transactions))
		} else {
			details = details[:0]
		}

		for _, txHash := range block.transactions {
			k := keyTxRecord(&txHash, &block.Block)
			v := existsRawTxRecord(ns, k)
			if v == nil {
				str := fmt.Sprintf("missing transaction %v for "+
					"block %v", txHash, block.Order)
				return false, storeError(ErrData, str, nil)
			}
			detail := TxDetails{
				Block: BlockMeta{
					Block: block.Block,
					Time:  block.Time,
				},
			}
			err := readRawTxRecord(&txHash, v, &detail.TxRecord)
			if err != nil {
				return false, err
			}

			credIter := makeReadCreditIterator(ns, k)
			for credIter.next() {
				if int(credIter.elem.Index) >= len(detail.MsgTx.TxOut) {
					str := "saved credit index exceeds number of outputs"
					return false, storeError(ErrData, str, nil)
				}

				// The credit iterator does not record whether
				// this credit was spent by an unMined
				// transaction, so check that here.
				if !credIter.elem.Spent {
					k := canonicalOutPoint(&txHash, credIter.elem.Index)
					spent := existsRawUnMinedInput(ns, k) != nil
					credIter.elem.Spent = spent
				}
				detail.Credits = append(detail.Credits, credIter.elem)
			}
			if credIter.err != nil {
				return false, credIter.err
			}

			debIter := makeReadDebitIterator(ns, k)
			for debIter.next() {
				if int(debIter.elem.Index) >= len(detail.MsgTx.TxIn) {
					str := "saved debit index exceeds number of inputs"
					return false, storeError(ErrData, str, nil)
				}

				detail.Debits = append(detail.Debits, debIter.elem)
			}
			if debIter.err != nil {
				return false, debIter.err
			}

			details = append(details, detail)
		}

		// Every block record must have at least one transaction, so it
		// is safe to call f.
		brk, err := f(details)
		if err != nil || brk {
			return brk, err
		}
	}
	return false, blockIter.err
}

// RangeTransactions runs the function f on all transaction details between
// blocks on the best chain over the height range [begin,end].  The special
// height -1 may be used to also include unMined transactions.  If the end
// height comes before the begin height, blocks are iterated in reverse order
// and unMined transactions (if any) are processed first.
//
// The function f may return an error which, if non-nil, is propagated to the
// caller.  Additionally, a boolean return value allows exiting the function
// early without reading any additional transactions early when true.
//
// All calls to f are guaranteed to be passed a slice with more than zero
// elements.  The slice may be reused for multiple blocks, so it is not safe to
// use it after the loop iteration it was acquired.
func (s *Store) RangeTransactions(ns walletdb.ReadBucket, begin, end int32,
	f func([]TxDetails) (bool, error)) error {

	var addedUnMined bool
	if begin < 0 {
		brk, err := s.rangeUnMinedTransactions(ns, f)
		if err != nil || brk {
			return err
		}
		addedUnMined = true
	}

	brk, err := s.rangeBlockTransactions(ns, begin, end, f)
	if err == nil && !brk && !addedUnMined && end < 0 {
		_, err = s.rangeUnMinedTransactions(ns, f)
	}
	return err
}

// PreviousPkScripts returns a slice of previous output scripts for each credit
// output this transaction record debits from.
func (s *Store) PreviousPkScripts(ns walletdb.ReadBucket, rec *TxRecord, block *Block) ([][]byte, error) {
	var pkScripts [][]byte

	if block == nil {
		for _, input := range rec.MsgTx.TxIn {
			prevOut := &input.PreviousOut

			v := existsRawUnMined(ns, prevOut.Hash[:])
			if v != nil {
				// Ensure a credit exists for this
				// unMined transaction before including
				// the output script.
				k := canonicalOutPoint(&prevOut.Hash, prevOut.OutIndex)
				if existsRawUnMinedCredit(ns, k) == nil {
					continue
				}

				pkScript, err := fetchRawTxRecordPkScript(
					prevOut.Hash[:], v, prevOut.OutIndex)
				if err != nil {
					return nil, err
				}
				pkScripts = append(pkScripts, pkScript)
				continue
			}

			_, credKey := existsUnspent(ns, prevOut)
			if credKey != nil {
				k := extractRawCreditTxRecordKey(credKey)
				v = existsRawTxRecord(ns, k)
				pkScript, err := fetchRawTxRecordPkScript(k, v,
					prevOut.OutIndex)
				if err != nil {
					return nil, err
				}
				pkScripts = append(pkScripts, pkScript)
				continue
			}
		}
		return pkScripts, nil
	}

	recKey := keyTxRecord(&rec.Hash, block)
	it := makeReadDebitIterator(ns, recKey)
	for it.next() {
		credKey := extractRawDebitCreditKey(it.cv)
		index := extractRawCreditIndex(credKey)
		k := extractRawCreditTxRecordKey(credKey)
		v := existsRawTxRecord(ns, k)
		pkScript, err := fetchRawTxRecordPkScript(k, v, index)
		if err != nil {
			return nil, err
		}
		pkScripts = append(pkScripts, pkScript)
	}
	if it.err != nil {
		return nil, it.err
	}

	return pkScripts, nil
}
