<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex">
        <el-col :span="6">
          <h2>导入/导出</h2>
        </el-col>
        <el-col :span="12"></el-col>
        <el-col :span="6">
          <el-button type="primary" icon="el-icon-plus" @click="importKey" size="small">导入私钥</el-button>
        </el-col>
      </el-row>
    </el-header>
    <el-main class="cmain">
      <el-table :data="tableData">
        <el-table-column prop="addr" label="地址" width="400px"></el-table-column>
        <el-table-column label>
          <template slot-scope="scope">
            <el-button @click="dumpKey(scope.row.addr)" type="text" icon="el-icon-download">导出私钥</el-button>
          </template>
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
    importKey() {
      this.$router.push({ path: "/backup/import" });
    },
    getAddressList() {
      let _this = this;
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "wallet_getAddressesByAccount",
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
    },
    checkWalletStats(callback) {
      let _this = this;
      _this.$emit("getWalletStats", stats => {
        if (stats != "unlock") {
          _this.$emit("walletPasswordDlg", "wallet_unlock", result => {
            if (!result) {
              // _this.$router.push("/");
              return;
            }
            callback();
          });
          return;
        }
        callback();
      });
    },
    dumpKey(addr) {
      let _this = this;

      const h = this.$createElement;

      let dumpKeyDo = () => {
        _this
          .$axios({
            method: "post",
            data: JSON.stringify({
              id: new Date().getTime(),
              method: "wallet_dumpPrivKey",
              params: [addr]
            })
          })
          .then(response => {
            if (typeof response.data.error != "undefined") {
              _this.$emit("alertResError", response.data.error, () => {});
              return;
            }
            let msg = h("div", [
              h("p", null, "地址: " + addr),
              h(
                "p",
                { style: "word-wrap:break-word" },
                "私钥: " + response.data.result
              )
            ]);
            this.$alert(msg, "私钥", {
              showClose: false,
              confirmButtonText: "确定",
              callback: () => {}
            });
          });
      };

      _this.checkWalletStats(dumpKeyDo);
    }
  }
};
</script>