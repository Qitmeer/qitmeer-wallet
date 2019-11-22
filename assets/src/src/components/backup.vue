<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex">
        <el-col :span="6">
          <h2>导入/导出</h2>
        </el-col>
        <el-col :span="12"></el-col>
        <el-col :span="6">
          <el-button type="primary" icon="el-icon-plus" @click="newAccount" size="small">导入私钥</el-button>
        </el-col>
      </el-row>
    </el-header>
    <el-main class="cmain">
      <el-table :data="tableData">
        <el-table-column prop="addr" label="地址"></el-table-column>
        <el-table-column label>
          <el-link type="primary" icon="el-icon-download">导出key</el-link>&nbsp;&nbsp;
        </el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    const item = [];
    return {
      tableData: item,
      currentAccount: "imported"
    };
  },
  mounted() {
    this.getAddressList();
  },
  methods: {
    newAccount() {
      this.$router.push({ path: "/account/new" });
    },
    getAddressList() {
      let _this = this;
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "account_listAddresses",
            params: [_this.currentAccount]
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

          _this.tableData = tmpTable;
        });
    }
  }
};
</script>