<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex" justify="space-between">
        <el-col :span="4">
          <h2>发送交易</h2>
        </el-col>
        <el-col :span="8">
          <el-row>
            <el-col :span="8">
              <span style="color:#303133;font-weight:600;">当前账户：</span>
            </el-col>
            <el-col :span="14">
              <el-select v-model="currentAccount" placeholder="账号" size="small">
                <el-option
                  v-for="item in accounts"
                  :key="item.account"
                  :label="item.account"
                  :value="item.account"
                ></el-option>
              </el-select>
            </el-col>
          </el-row>
        </el-col>
        <el-col :span="4"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-form :model="form" label-width="120px">
        <el-form-item label="可用余额">
          <el-input v-model="currentAccountBalance" :disabled="true"></el-input>
        </el-form-item>
        <el-form-item label="转给(to)">
          <el-input placeholder="address" v-model="form.to"></el-input>
        </el-form-item>
        <el-form-item label="转账金额(amount)">
          <el-input placeholder="amount" v-model="form.value"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="sentTx">创建</el-button>
          <el-button @click="toTxList">取消</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    return {
      accounts: [],
      currentAccount: "",
      currentAccountBalance: 0,
      form: {
        to: "",
        value: ""
      }
    };
  },
  mounted() {
    let _this = this;
    _this.$emit("getWalletStats", stats => {
      if (stats != "unlock") {
        _this.$emit("walletPasswordDlg", "wallet_unlock", result => {
          if (!result) {
            _this.$router.push("/");
          }
        });
      }
    });

    if (this.$store.state.Accounts.length == 0) {
      this.$router.push("/account");
      return;
    }

    this.accounts = this.$store.state.Accounts;
    this.currentAccount = this.$store.state.Accounts[0].account;
    this.currentAccountBalance = this.$store.state.Accounts[0].UnspendAmount;
  },
  methods: {
    toTxList() {
      this.$router.push("/tx/list");
    },
    sentTx() {
      let _this = this;
      //
      //SendToAddress
      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "wallet_sendToAddressByAccount",
          params: [
            _this.currentAccount,
            _this.form.to,
            parseFloat(_this.form.value),
            "",
            ""
          ]
        })
      }).then(response => {
        const h = this.$createElement;
        if (typeof response.data.error != "undefined") {
          _this.$emit("alertResError", response.data.error, () => {});
          return;
        }
        this.$alert("发送交易成功", {
          showClose: false,
          closeOnClickModal: false,
          closeOnPressEscape: false,
          confirmButtonText: "确定",
          callback: (action, instance) => {
            this.$router.push({ path: "/tx/list" });
          }
        });
      });
    }
  },
  watch: {
    currentAccount() {
      for (let i = 0; i < this.accounts.length; i++) {
        if (this.accounts[i].account == this.currentAccount) {
          this.currentAccount = this.accounts[i].account;
          this.currentAccountBalance = this.accounts[i].UnspendAmount;
          return;
        }
      }
    }
  }
};
</script>