<template>
  <div
    id="app"
    v-loading="loading"
    v-bind:element-loading-text="loadingText"
    element-loading-spinner="el-icon-loading"
    element-loading-background="rgba(0, 0, 0, 0.8)"
  >
    <div id="sidebar">
      <div id="logo">
        <!-- <img src="./assets/logo.png" style="with:40px;height:40px;" /> -->
        <h1>
          <router-link to="/">QITMEER WALLET</router-link>
        </h1>
      </div>
      <div id="menu">
        <h2>用户</h2>
        <router-link to="/account">
          <i class="iconfont icon-Account"></i>&nbsp;
          账户管理
        </router-link>
        <router-link to="/address" icon="el-icon-tickets">
          <i class="iconfont icon-qianbao"></i>&nbsp;
          地址管理
        </router-link>
        <router-link to="/tx/send">
          <i class="iconfont icon-jiaoyi"></i>&nbsp;
          发送交易
        </router-link>
        <router-link to="/tx/list">
          <i class="iconfont icon-zhangdan"></i>&nbsp;
          交易记录
        </router-link>
        <router-link to="/backup">
          <i class="iconfont icon-baobeiguanli-baobeibeifen"></i>&nbsp;
          备份/恢复
        </router-link>

        <h2>数据</h2>
        <router-link to="/node">
          <i class="iconfont icon-jiedian"></i>&nbsp;
          节点管理
        </router-link>
        <!-- <a href="#">块数据查询</a> -->
      </div>
      <div id="info">
        <h2>
          当前节点: {{nodeName}}
          <el-dropdown>
            <span class="el-dropdown-link">
              <i class="el-icon-s-tools"></i>
            </span>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item v-for="(item,index) in nodes" :key="index">{{item}}</el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </h2>
        <p>
          <span>网络:</span>Mainnet
        </p>
        <p>
          <span>全网:</span>100000
        </p>
        <p>
          <span>节点:</span>1000
        </p>
      </div>
      <div class="version">
        <p>
          <span>版本:</span>201907-0.0.1
        </p>
      </div>
    </div>
    <router-view
      class="mainwarp"
      @setLoading="setLoading"
      @checkWalletStats="checkWalletStats"
      @dlgUnlockWallet="dlgUnlockWallet"
      @alertResError="alertResError"
      @updateAccounts="updateAccounts"
    ></router-view>
  </div>
</template>

<style>
</style>

<script>
export default {
  name: "app",
  data() {
    var nodes = ["loacl", "dao"];

    return {
      nodeName: "local",
      nodes: nodes,
      loading: false,
      loadingText: "",
      walletcrate: true,
      openForm: {
        walletPass: ""
      }
    };
  },
  mounted() {},
  methods: {
    setLoading(b, txt) {
      this.loading = b;
      this.loadingText = txt;
    },
    wathStats() {},
    checkWalletStats(callback) {
      let _this = this;
      _this.loading = true;
      _this.loadingText = "检查钱包状态";
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "wallet_status",
            params: null
          })
        })
        .then(response => {
          if (typeof response.data.error != "undefined") {
            _this.$message({
              message: "获取钱包信息异常，请刷新! /n " + response.data.error,
              type: "warning",
              duration: 500,
              onClose: function() {
                _this.$router.push("/");
              }
            });
            return;
          }
          _this.loading = false;
          _this.loadingText = "";

          let rs = response.data.result;
          switch (rs.stats) {
            case "nil":
              if (_this.$route.path == "/") {
                _this.dlgCreateWallet();
              }
              if (
                !(
                  _this.$route.path == "/wallet/create" ||
                  _this.$route.path == "/wallet/recove"
                )
              ) {
                _this.$router.push("/");
              }
              break;
            case "closed":
              _this.openWallet();
              break;
            case "lock":
              callback("lock");
              break;
            case "unlock":
              callback("unlock");
              // console.log("opend", _this.$route.path);
              // if (
              //   _this.$route.path == "/wallet/create" ||
              //   _this.$route.path == "/wallet/recove"
              // ) {
              //   _this.$router.push("/");
              // }
              //
              //this.$store.state.accounts = response.data.result;
              break;
          }
        });
    },
    openWallet() {
      let _this = this;
      this.loading = true;
      this.loadingText = "打开钱包";
      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "wallet_open",
          params: [this.openForm.walletPass]
        })
      }).then(response => {
        this.loading = false;
        if (typeof response.data.error != "undefined") {
          this.$message({
            message: "打开钱包错误：" + response.data.error.message,
            type: "warning",
            duration: 500,
            onClose: function() {
              _this.$emit("setLoading", false);
              _this.$router.go(0);
            }
          });
          return;
        }
        console.log("open wallet ok!");
        this.$router.push("/");
      });
    },
    dlgCreateWallet() {
      this.$confirm("", "创建新钱包", {
        showClose: false,
        closeOnClickModal: false,
        confirmButtonText: "创建新钱包",
        cancelButtonText: "恢复钱包",
        callback: (action, instance) => {
          if (action == "cancel") {
            this.$router.push({ path: "/wallet/recove" });
          } else {
            this.$router.push({ path: "/wallet/create" });
          }
        }
      });
    },
    dlgUnlockWallet(callback) {
      let _this = this;
      _this
        .$prompt("解锁钱包", {
          confirmButtonText: "确定",
          cancelButtonText: "取消"
        })
        .then(({ value }) => {
          this.$axios({
            method: "post",
            data: JSON.stringify({
              id: new Date().getTime(),
              method: "wallet_unlock",
              params: [value, 2000]
            })
          }).then(response => {
            if (typeof response.data.error != "undefined") {
              this.$message({
                message: h("div", null, [
                  h("p", null, "错误！"),
                  h("p", null, "code:" + response.data.error.code),
                  h("p", null, "info:" + response.data.error.message)
                ]),
                type: "warning",
                duration: 500,
                onClose: function() {}
              });
              callback(false);
              return;
            }
            callback(true);
          });
        })
        .catch(() => {
          callback(false);
        });
    },
    alertResError(error, callbackAction) {
      const h = this.$createElement;
      this.$alert(
        h("div", null, [
          h("p", null, "错误！"),
          h("p", null, "code:" + error.code),
          h("p", null, "info:" + error.message)
        ]),
        {
          showClose: false,
          closeOnClickModal: false,
          closeOnPressEscape: false,
          confirmButtonText: "确定",
          callback: callbackAction
        }
      );
    },
    listAccount2table(listAccounts) {
      let tmpTable = [];
      for (let k in listAccounts) {
        tmpTable.push({
          account: k,
          balance: listAccounts[k]
        });
      }
      return tmpTable;
    },
    updateAccounts(callback) {
      let _this = this;
      _this.loading = true;
      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "account_list",
          params: null
        })
      }).then(response => {
        _this.loading = false;
        if (typeof response.data.error != "undefined") {
          _this.alertResError(response.data.error, () => {
            _this.push("/");
          });
          return;
        }
        _this.$store.state.Accounts = _this.listAccount2table(
          response.data.result
        );

        callback();
      });
    }
  },
  watch: {
    needOneAccount: function() {
      if (this.$store.state.accounts.length == 0) {
        return false;
      }
    }
  }
};
</script>


