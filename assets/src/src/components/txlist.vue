<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex" justify="space-between">
        <el-col :span="4">
          <h2>交易记录</h2>
        </el-col>
        <el-col :span="20">
          <el-row>
            <el-col :span="2">
              <span style="color:#303133;font-weight:600;">账户：</span>
            </el-col>
            <el-col :span="4">
              <el-select v-model="currentAccount" placeholder="账号" size="small">
                <el-option
                    v-for="item in accounts"
                    :key="item.account"
                    :label="item.account"
                    :value="item.account"
                ></el-option>
              </el-select>
            </el-col>
            <el-col :span="2">
              <span style="color:#303133;font-weight:600;">地址：</span>
            </el-col>
            <el-col :span="14">
              <el-select v-model="currentAddress" style="width:100%">
                <el-option
                    v-for="item in addresses"
                    :key="item.addr"
                    :label="item.addr"
                    :value="item.addr"
                ></el-option>
              </el-select>
            </el-col>
          </el-row>
        </el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-table :data="txList">
        <el-table-column prop="txId" label="txId" width="470"></el-table-column>
        <el-table-column prop="type" label="in/out" width="60"></el-table-column>
        <el-table-column prop="amount" label="金额"></el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    // var txlist = [
    //   {
    //     type: "in",
    //     from: "asdfasdfasdfasdf",
    //     to: "asdfasdfasdfasdf",
    //     amount: 34234,
    //     date: "2019-07-23 23:32:23"
    //   }
    // ];

    return {
      txList: [],
      accounts: [],
      addresses: ["*"],
      currentAccount: "",
      currentAddress: "*"
    };
  },
  mounted() {
    if (this.$store.state.Accounts.length == 0) {
      this.$router.push("/account");
      return;
    }
    this.accounts = this.$store.state.Accounts;
    this.currentAccount = this.$store.state.Accounts[0].account;
  },
  methods: {
    bills2table(bills) {
      let _this = this;
      let tmpTable = [];
      for (let i = 0; i < bills.length; i++) {
        let bill = bills[i]
        let amount = bill.amount / Math.pow(10.0, 8);
        let inOut = amount <= 0 ? "OUT" : "IN";

        tmpTable.push({
          date: "",
          type: inOut,
          txId: bill.tx_id,
          amount: amount
        });
      }
      _this.txList = tmpTable;
    },
    getBills(addr) {
      if (addr == "*") {
        //todo all account addresses
        return;
      }

      let _this = this;

      _this.$emit("setLoading", true, "");

      _this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "wallet_getBillsByAddr",
          params: [addr, 2, -1, 100]
        })
      }).then(resp => {
        _this.$emit("setLoading", false, "");
        if (typeof resp.data.error != "undefined") {
          _this.$emit("alertResError", resp.data.error, () => {
          });
          return;
        }

        _this.bills2table(resp.data.result.bills);
      }).catch(() => {
        _this.$emit("setLoading", false, "");
      });
    },
    getAddressList(account) {
      let _this = this;
      _this
          .$axios({
            method: "post",
            data: JSON.stringify({
              id: new Date().getTime(),
              method: "wallet_getAddressesByAccount",
              params: [account]
            })
          })
          .then(response => {
            if (typeof response.data.error != "undefined") {
              _this.$emit("alertResError", response.data.error, () => {
              });
              return;
            }

            let tmpTable = [];
            for (let i = 0; i < response.data.result.length; i++) {
              tmpTable.push({addr: response.data.result[i]});
            }
            _this.addresses = tmpTable;
            _this.currentAddress = _this.addresses[0].addr;
          });
    }
  },
  watch: {
    currentAccount() {
      this.getAddressList(this.currentAccount);
    },
    currentAddress() {
      this.txList = [];
      this.getBills(this.currentAddress);
    }
  }
};
</script>