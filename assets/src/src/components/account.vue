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
        <el-table-column prop="UnspendAmount" label="余额(可花费)"></el-table-column>
        <el-table-column prop="ConfirmAmount" label="余额(待确认)"></el-table-column>
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
      let tmpTable = [];
      for (let k in listAccounts) {
        tmpTable.push({
          account: k,
          UnspendAmount: listAccounts[k].UnspendAmount / 100000000,
          ConfirmAmount: listAccounts[k].ConfirmAmount / 100000000
        });
      }
      return tmpTable;
    },
    newAccount() {
      this.$router.push({ path: "/account/new" });
    }
  },
  mounted() {
    let _this = this;
    this.$axios({
      method: "post",
      data: JSON.stringify({
        id: new Date().getTime(),
        method: "wallet_getAccountsAndBalance",
        params: null
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