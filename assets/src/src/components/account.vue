<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex">
        <el-col :span="6">
          <h2>账号管理</h2>
        </el-col>
        <el-col :span="12"></el-col>
        <el-col :span="6">
          <el-button type="primary" icon="el-icon-plus" @click="newAccount" size="small">新建账号</el-button>
        </el-col>
      </el-row>
    </el-header>
    <el-main class="cmain">
      <el-table :data="accountsTable" :key="Math.random()">
        <el-table-column prop="account" label="名称" width="120"></el-table-column>
        <el-table-column prop="UnspentAmount" label="余额(可花费)"></el-table-column>
        <el-table-column prop="ConfirmAmount" label="余额(待确认)"></el-table-column>
        <el-table-column prop="LockAmount" label="余额(锁定)"></el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    return {
      accountsTable: [
        {
          account: "defalut",
          balance: 0
        }
      ]
    };
  },
  methods: {
    listAccount2table(listAccounts) {
      // eslint-disable-next-line no-console
      console.log(listAccounts)
      let tmpTable = [];
     /* let i = 0;
      // eslint-disable-next-line no-unused-vars
      for (let item in listAccounts) {
        // if (!item.hasOwnProperty(listAccounts)) return;
        tmpTable.push({
          account: i,
          UnspendAmount: listAccounts[item]['UnspentAmount']['Value'] / 1e8,
          ConfirmAmount: listAccounts[item]['UnconfirmedAmount']['Value'] / 1e8,
        })
      }*/

      for (let k in listAccounts) {
        // eslint-disable-next-line no-console
        console.log(k)
         tmpTable.push({
           account: k,
           UnspentAmount: listAccounts[k]['UnspentAmount'] / 1e8,
           LockAmount: listAccounts[k]['LockAmount'] / 1e8,
           ConfirmAmount: listAccounts[k]['UnconfirmedAmount'] / 1e8,
         });
       }
      return tmpTable;
    },
    newAccount() {
      this.$router.push({path: "/account/new"});
    }
  },
  mounted() {
    let _this = this;
    this.$axios({
      method: "post",
      data: JSON.stringify({
        id: new Date().getTime(),
        method: "wallet_getAccountsAndBalance",
        params: ["MEER"]
      })
    }).then(response => {
      if (typeof response.data.error != "undefined") {
        _this.$emit("alertResError", response.data.error, () => {
          _this.$router.push("/");
        });
        return;
      }
      _this.accountsTable = _this.listAccount2table(response.data.result);
      _this.$store.state.Accounts = _this.accountsTable;
    });
  }
};
</script>