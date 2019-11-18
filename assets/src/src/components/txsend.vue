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
              <el-select v-model="current" placeholder="账号" size="small">
                <el-option
                  v-for="item in accounts"
                  :key="item.index"
                  :label="item.account"
                  :value="item.index"
                  @change="changeAccount"
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
          <el-input v-model="balance" :disabled="true"></el-input>
        </el-form-item>
        <el-form-item label="转给(to)">
          <el-input placeholder="address" v-model="form.to"></el-input>
        </el-form-item>
        <el-form-item label="转账金额(amount)">
          <el-input placeholder="amount" v-model="form.value"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="sentTx">创建</el-button>
          <el-button @click="toAccount">取消</el-button>
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
      current: "",
      balance: "",
      form: {
        to: "",
        value: ""
      }
    };
  },
  mounted() {
    if (this.$store.state.Accounts.length == 0) {
      this.$router.push("/account");
      return;
    }
    this.accounts = this.$store.state.Accounts;
    this.current = this.$store.state.Accounts[0].account;
    this.balance = this.$store.state.Accounts[0].balance;
  },
  methods: {
    toAccount() {
      this.$router.push("/account");
    },
    changeAccount(index) {
      this.balance = this.$store.state.Accounts[index].balance;
    },
    sentTx() {
      let _this = this;
      //
      //SendToAddress
      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "account_sendToAddress",
          params: [_this.form.to, parseFloat(_this.form.value), "", ""]
        })
      }).then(response => {
        const h = this.$createElement;
        if (typeof response.data.error != "undefined") {
          _this.$emit("alertResError", response.data.error, () => {});
          return;
        }
        this.$alert("创建账号成功", {
          showClose: false,
          closeOnClickModal: false,
          closeOnPressEscape: false,
          confirmButtonText: "确定",
          callback: (action, instance) => {
            this.$router.push({ path: "/account" });
          }
        });
      });
    }
  }
};
</script>