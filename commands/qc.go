package commands

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/qx"
	"github.com/spf13/cobra"
	"strconv"
)

var QcCmd = &cobra.Command{
	Use:   "qc",
	Short: "qitmeer wallet command",
	Long:  `qitmeer wallet command`,
}

func AddQcCommand() {
	QcCmd.AddCommand(createWalletCmd)
	QcCmd.AddCommand(setSynceToNumCmd)
	QcCmd.AddCommand(createNewAccountCmd)
	QcCmd.AddCommand(getnewaddressCmd)
	QcCmd.AddCommand(getBalanceCmd)
	QcCmd.AddCommand(newGetListTxByAddrCmd())
	QcCmd.AddCommand(newGetBillByAddrCmd())
	QcCmd.AddCommand(updateblockCmd)
	QcCmd.AddCommand(syncheightCmd)
	QcCmd.AddCommand(sendToAddressCmd)
	QcCmd.AddCommand(newImportPrivKeyCmd())
	QcCmd.AddCommand(getAddressesByAccountCmd)
	QcCmd.AddCommand(listAccountsBalanceCmd)
	QcCmd.AddCommand(newGetTxByTxIdCmd())
	QcCmd.AddCommand(getTxSpendInfoCmd)
}

var createWalletCmd = &cobra.Command{
	Use:     "create",
	Short:   "create wallet",
	Example: "create",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		CreatWallet()
	},
}

var createNewAccountCmd = &cobra.Command{
	Use:     "createnewaccount {account} {pripassword}",
	Short:   "create new account",
	Example: "createnewaccount test password",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = UnLock(args[1])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		_ = createNewAccount(args[0])
	},
}
var getnewaddressCmd = &cobra.Command{
	Use:     "getnewaddress {account}",
	Short:   "create new address by account",
	Example: "getnewaddress default",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		getNewAddress(args[0])
	},
}

var getBalanceCmd = &cobra.Command{
	Use:   "getbalance {address} {string ,company : i(int64),f(float),default i } {bool ,detail : true,false,default false }",
	Short: "getbalance",
	Example: `
		getbalance TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5	i true
		getbalance TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5	f false
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		company := "i"
		detail := "false"
		b, err := getBalance(args[0])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(args) > 1 {
			if args[1] != "i" {
				company = "f"
			}
			if len(args) > 2 {
				detail = args[2]
			}
		}
		if company == "i" {
			if detail == "true" {
				fmt.Printf("unspend:%s\n", b.UnspendAmount.String())
				fmt.Printf("unconfirmed:%s\n", b.ConfirmAmount.String())
				fmt.Printf("totalamount:%s\n", b.TotalAmount.String())
				fmt.Printf("spendamount:%s\n", b.SpendAmount.String())
			} else {
				fmt.Printf("%s\n", b.UnspendAmount.String())
			}
		} else {
			if detail == "true" {
				fmt.Printf("unspend:%f\n", b.UnspendAmount.ToCoin())
				fmt.Printf("unconfirmed:%f\n", b.ConfirmAmount.ToCoin())
				fmt.Printf("totalamount:%f\n", b.TotalAmount.ToCoin())
				fmt.Printf("spendamount:%f\n", b.SpendAmount.ToCoin())
			} else {
				fmt.Printf("%f\n", b.UnspendAmount.ToCoin())
			}
		}

	},
}

var setSynceToNumCmd = &cobra.Command{
	Use:   "setsyncetonum {num}  ",
	Short: "please use caution when specifying how many blocks to update from ",
	Example: `
		setsyncetonum 100000
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		order, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Error("setsyncetonum ", "error", err.Error())
			return
		}
		if order < 0 {
			log.Error("setsyncetonum ", "error", "The specified order cannot be less than 0")
			return
		}
		err = SetSynceToNum(order)
		if err == nil {
			fmt.Println("succ")
			return
		}
	},
}

var getTxSpendInfoCmd = &cobra.Command{
	Use:   "gettxspendinfo {txId} {index}",
	Short: "gettxspendinfo",
	Example: `
		gettxspendinfo 10c710ffcdf3bea9a21656c26fc0dd5796cb3d0b60aafb2ede49ca1248e9aa0d
		gettxspendinfo 10c710ffcdf3bea9a21656c26fc0dd5796cb3d0b60aafb2ede49ca1248e9aa0d	0
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		b, err := GetTxSpendInfo(args[0])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(args) > 1 {
			index, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Error("Argument is not of type int")
				return
			}
			if len(b) < int(index+1) {
				log.Error("Index out of array range")
				return
			} else {
				if b[index].SpendTo == nil {
					fmt.Printf("addr:%v,txid:%v,index:%v,unspend\n", b[index].Address, b[index].TxId, b[index].Index)
				} else {
					fmt.Printf("addr:%v,txid:%v,index:%v,spend to: txid:%v,index:%v\n", b[index].Address, b[index].TxId, b[index].Index, b[index].SpendTo.TxHash, b[index].SpendTo.Index)
				}
				return
			}
		} else {
			for _, output := range b {
				if output.SpendTo == nil {
					fmt.Printf("addr:%v,txid:%v,index:%v,unspend\n", output.Address, output.TxId, output.Index)
				} else {
					fmt.Printf("addr:%v,txid:%v,index:%v,spendto: txid:%v,index:%v\n", output.Address, output.TxId, output.Index, output.SpendTo.TxHash, output.SpendTo.Index)
				}
				return
			}
		}

	},
}

var sendToAddressCmd = &cobra.Command{
	Use:   "sendtoaddress {address} {amount} {pripassword} ",
	Short: "send transaction ",
	Example: `
		sendtoaddress TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5 10 pripassword
		`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		f32, err := strconv.ParseFloat(args[1], 32)
		if err != nil {
			log.Error("sendtoaddress ", "error", err.Error())
			return
		}
		err = UnLock(args[2])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		sendToAddress(args[0], float64(f32))
	},
}

func newGetTxByTxIdCmd() *cobra.Command {
	getTxByTxIdCmd := &cobra.Command{
		Use:   "gettx {txid}",
		Short: "Access to transaction information ",
		Example: `
		gettx 81278a6ba67d4ea2fc49fb469f2a45f6adb2306b82146747b9d5f3bd655e5030
		`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := OpenWallet(); err != nil {
				return err
			}

			_, err := getTx(args[0])
			return err
		},
	}

	return getTxByTxIdCmd
}

var getAddressesByAccountCmd = &cobra.Command{
	Use:   "getaddressesbyaccount {string ,account,defalut imported} ",
	Short: "get addresses by account ",
	Example: `
		getaddressesbyaccount imported
		`,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		account := "imported"
		if len(args) > 0 {
			account = args[0]
		}
		getAddressesByAccount(account)
	},
}

func newImportPrivKeyCmd() *cobra.Command {
	var format string
	importPrivKeyCmd := &cobra.Command{
		Use:   "importprivkey {priKey} {pripassword}",
		Short: "import priKey ",
		Example: `
		importprivkey  ef235aacf90d9f4aadd8c92e4b2562e1d9eb97f0df9ba3b508258739cb013db2 pripassword 
		importprivkey  5HueCGU8rMjxEXxiPuD5BDku4MkFqeZyd4dZ1jvhTVqvbTLvyTJ pripassword  --format=wif
		`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := OpenWallet(); err != nil {
				return err
			}
			if err := UnLock(args[1]); err != nil {
				return err
			}
			priv := args[0]
			if format == "wif" {
				if decoded, _, err := qx.DecodeWIF(priv); err != nil {
					return err
				} else {
					priv = hex.EncodeToString(decoded)
				}
			}

			_, err := importPrivKey(priv)
			return err
		},
	}

	importPrivKeyCmd.Flags().StringVarP(
		&format, "format", "f", "raw", "Private Key format. {raw, wif}")

	return importPrivKeyCmd
}

var listAccountsBalanceCmd = &cobra.Command{
	Use:   "listaccountsbalance ",
	Short: "list Accounts Balance",
	Example: `
		listaccountsbalance
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		listAccountsBalance()
	},
}

func newGetListTxByAddrCmd() *cobra.Command {
	filterFlag := "all"
	pageNoFlag := wallet.PageUseDefault
	pageSizeFlag := wallet.PageDefaultSize

	getListTxAddrCmd := &cobra.Command{
		Use:   "getlisttxbyaddr {address}",
		Short: "get all transactions by address",
		Long: `request all the transactions that affect a specific address, 
a transaction could affect MULTIPLE addresses`,
		Example: `
		getlisttxbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 --filter=in
		getlisttxbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 --filter=out 
		getlisttxbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 
		getlisttxbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 --page_no=1 --page_size=10
		`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := OpenWallet(); err != nil {
				return err
			}
			filter := wallet.FilterAll

			if filterFlag == "in" {
				filter = wallet.FilterIn
			} else if filterFlag == "out" {
				filter = wallet.FilterOut
			}

			_, err := getListTxByAddr(args[0], filter, pageNoFlag, pageSizeFlag)
			return err
		},
	}

	getListTxAddrCmd.Flags().StringVarP(
		&filterFlag, "filter", "f", "all", "Filter. {in, out, all}")
	getListTxAddrCmd.Flags().IntVarP(
		&pageNoFlag, "page_no", "i", wallet.PageUseDefault, "Page number.")
	getListTxAddrCmd.Flags().IntVarP(
		&pageSizeFlag, "page_size", "s", wallet.PageDefaultSize, "Page size.")

	return getListTxAddrCmd
}

func newGetBillByAddrCmd() *cobra.Command {
	filterFlag := "all"
	pageNoFlag := wallet.PageUseDefault
	pageSizeFlag := wallet.PageDefaultSize

	getBillAddrCmd := &cobra.Command{
		Use:   "getbillbyaddr {address}",
		Short: "get bill by address",
		Long: `request the bill of a specific address, a bill is the log of payments, 
which are the effect that a transaction makes on a specific address`,
		Example: `
		getbillbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 --filter=in
		getbillbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 --filter=out 
		getbillbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 
		getbillbyaddr Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7 --page_no=1 --page_size=10
		`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := OpenWallet(); err != nil {
				return err
			}
			filter := wallet.FilterAll

			if filterFlag == "in" {
				filter = wallet.FilterIn
			} else if filterFlag == "out" {
				filter = wallet.FilterOut
			}

			_, err := getBillByAddr(args[0], filter, pageNoFlag, pageSizeFlag)
			return err
		},
	}

	getBillAddrCmd.Flags().StringVarP(
		&filterFlag, "filter", "f", "all", "Filter. {in, out, all}")
	getBillAddrCmd.Flags().IntVarP(
		&pageNoFlag, "page_no", "i", wallet.PageUseDefault, "Page number.")
	getBillAddrCmd.Flags().IntVarP(
		&pageSizeFlag, "page_size", "s", wallet.PageMaxSize, "Page size.")

	return getBillAddrCmd
}

var updateblockCmd = &cobra.Command{
	Use:   "updateblock {int,Update to the specified block, 0 is updated to the latest by default,defalut 0}",
	Short: "Update local block data",
	Example: `
		updateblock
		updateblock 12
		`,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		var height = int64(0)
		if len(args) > 0 {
			var err error
			height, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Info("Argument is not of type int")
				return
			}
		}
		updateblock(height)
	},
}
var syncheightCmd = &cobra.Command{
	Use:   "syncheight",
	Short: "Get the number of local synchronization blocks",
	Example: `
		syncheight
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		syncheight()
	},
}
