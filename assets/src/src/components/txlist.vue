<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex" justify="space-between">
        <el-col :span="4">
          <h2>交易记录</h2>
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
                ></el-option>
              </el-select>
            </el-col>
          </el-row>
        </el-col>
        <el-col :span="4"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-table :data="txList">
        <el-table-column prop="date" label="时间" width="160"></el-table-column>
        <el-table-column prop="type" label="-" width="40"></el-table-column>
        <el-table-column prop="to" label="to" width="240"></el-table-column>
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
      current: ""
    };
  },
  mounted() {
    if (this.$store.state.Accounts.length == 0) {
      this.$router.push("/account");
      return;
    }
    this.accounts = this.$store.state.Accounts;
    this.current = this.$store.state.Accounts[0].account;

    this.getTxList();
  },
  methods: {
    getTxList() {
      return [];

      let _this = this;
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "account_listTxs",
            params: [_this.current]
          })
        })
        .then(response => {
          if (typeof response.data.error != "undefined") {
            _this.$emit("alertResError", response.data.error, () => {});
            return;
          }

          let tmpTable = [];
          for (let i = 0; i < response.data.result.length; i++) {
            tmpTable.push({ addr: response.data.result[i] });
          }

          _this.txList = tmpTable;
        });
    }
  }
};
</script>