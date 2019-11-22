<template>
  <div
    id="app"
    v-loading.fullscreen.lock="loading"
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
          私钥导入
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
          节点: {{qitmeerdStatus.CurrentName}}
          <el-dropdown @command="qitmeerdChange">
            <span class="el-dropdown-link">
              <i class="el-icon-s-tools"></i>
            </span>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item
                v-for="(item,i) in qitmeerdList"
                :key="i"
                :command="item.Name"
              >{{item.Name}}</el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </h2>
        <p>网络: {{qitmeerdStatus.Network}}</p>
        <p>
          <span>MainOrder: {{qitmeerdStatus.MainOrder}}</span>
        </p>
        <p>
          <span>MainHeight: {{qitmeerdStatus.MainHeight}}</span>
        </p>
        <p>挖矿难度</p>
        <p>
          <span>
            Blake2bd:
            {{qitmeerdStatus.Blake2bdDiff}}
          </span>
        </p>
        <p>
          <span>Cuckaroo: {{qitmeerdStatus.CuckarooDiff}}</span>
        </p>
        <p>
          <span>Cuckatoo: {{qitmeerdStatus.CuckatooDiff}}</span>
        </p>
      </div>
      <div class="version">
        <p>
          <!-- <span>版本:</span>201907-0.0.1 -->
        </p>
      </div>
    </div>
    <router-view
      class="mainwarp"
      @setLoading="setLoading"
      @createWalletDlg="createWalletDlg"
      @getWalletStats="getWalletStats"
      @walletPasswordDlg="walletPasswordDlg"
      @walletOk="walletOk"
      @alertResError="alertResError"
      @updateAccounts="updateAccounts"
      @getQitmeerdList="getQitmeerdList"
    ></router-view>
    <el-dialog
      :visible.sync="walletPassword.dlgVisible"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      :v-loading="walletPassword.loading"
    >
      <el-form>
        <el-form-item :label="walletPassword.title">
          <el-input v-model="walletPassword.password" show-password></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="walletPasswordDlgGo(false)">取 消</el-button>
        <el-button type="primary" @click="walletPasswordDlgGo(true)">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<style>
</style>

<script>
export default {
  name: "app",
  data() {
    var qitmeerdList = { Local: {} };

    return {
      walletStatus: "unknown",
      qitmeerdList: qitmeerdList,
      qitmeerdStatusAlert: null,
      qitmeerdStatus: {
        Network: "",
        CurrentName: "Local",
        MainOrder: "",
        MainHeight: "",
        Blake2bdDiff: "",
        CuckarooDiff: "",
        CuckatooDiff: ""
      },
      loading: false,
      loadingText: "",
      walletcrate: true,
      openForm: {
        walletPass: ""
      },
      unlockWallet: {
        dlgVisible: false,
        pass: "",
        callback: () => {}
      },
      walletPassword: {
        title: "请输入登录密码",
        method: "wallet_open",
        dlgVisible: false,
        password: "",
        loading: false,
        unlockTimeout: 5 * 60, //wallet_unlock default 5 minutes
        callback: () => {}
      }
    };
  },
  mounted() {},
  methods: {
    setLoading(b, txt) {
      this.loading = b;
      this.loadingText = txt;
    },
    walletOk() {
      this.getQitmeerdStatus();
      this.getQitmeerdList(() => {});
      setInterval(this.getQitmeerdStatus, 1000 * 30);
    },
    getWalletStats(callback) {
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

          //let rs = response.data.result;
          callback(response.data.result.stats);
          _this.$store.state.Wallet = response.data.result.stats;
          // switch (rs.stats) {
          //   case "nil":
          //     if (_this.$route.path == "/") {
          //       _this.dlgCreateWallet();
          //     }
          //     if (
          //       !(
          //         _this.$route.path == "/wallet/create" ||
          //         _this.$route.path == "/wallet/recover"
          //       )
          //     ) {
          //       _this.$router.push("/");
          //     }
          //     break;
          //   case "closed":
          //     _this.openWallet();
          //     break;
          //   case "lock":
          //     callback("lock");
          //     break;
          //   case "unlock":
          //     callback("unlock");
          //     break;
          //   default:
          //     callback("unknown");
          //     break;
          // }
        })
        .catch(() => {
          callback("error");
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
    createWalletDlg() {
      let _this = this;
      _this.$confirm("", "创建新钱包", {
        showClose: false,
        closeOnClickModal: false,
        closeOnPressEscape: false,
        confirmButtonText: "创建新钱包",
        cancelButtonText: "恢复钱包",
        callback: (action, instance) => {
          if (action == "cancel") {
            _this.$router.push({ path: "/wallet/recover" });
          } else {
            _this.$router.push({ path: "/wallet/create" });
          }
        }
      });
    },
    walletPasswordDlg(method, callback) {
      let _this = this;
      if (method == "wallet_unlock") {
        _this.walletPassword.title = "请输入交易密码";
      } else {
        _this.walletPassword.title = "请输入登录密码";
      }
      _this.walletPassword.password = "";
      _this.walletPassword.method = method;
      _this.walletPassword.callback = callback;
      _this.walletPassword.dlgVisible = true;
    },
    walletPasswordDlgGo(go) {
      let _this = this;
      if (!go) {
        _this.walletPassword.dlgVisible = false;
        _this.walletPassword.callback(false);
        return;
      }
      _this.loading = true;

      let params = [_this.walletPassword.password];
      if (_this.walletPassword.method == "wallet_unlock") {
        params.push(_this.walletPassword.unlockTimeout);
      }

      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: _this.walletPassword.method, //"wallet_open","wallet_unlock"
            params: params
          })
        })
        .then(response => {
          _this.loading = false;
          if (typeof response.data.error != "undefined") {
            _this.alertResError(response.data.error, () => {
              _this.walletPassword.callback(false);
            });
          } else {
            _this.walletPassword.dlgVisible = false;
            _this.walletPassword.callback(true);
          }
        })
        .catch(() => {
          _this.loading = false;
          _this.walletPassword.callback(false);
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
    },
    getQitmeerdList(callback) {
      let _this = this;
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "qitmeerd_list",
            params: null
          })
        })
        .then(response => {
          if (typeof response.data.error != "undefined") {
            _this.$message({
              message: "获取节点信息异常，请刷新! /n " + response.data.error,
              type: "warning",
              duration: 500,
              onClose: function() {}
            });
            return;
          }
          let rs = response.data.result;
          _this.qitmeerdList = rs;
          _this.$store.state.QitmeerdList = rs;
          callback();
        });
    },
    qitmeerdChange(name) {
      let _this = this;
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "qitmeerd_reset",
            params: [name]
          })
        })
        .then(response => {
          if (typeof response.data.error != "undefined") {
            _this.$message({
              message: "节点信息异常，请刷新! /n " + response.data.error,
              type: "warning",
              duration: 5000,
              onClose: function() {}
            });
            return;
          }
          _this.getQitmeerdStatus();
        });
    },
    getQitmeerdStatus() {
      let _this = this;
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "qitmeerd_status",
            params: null
          })
        })
        .then(response => {
          if (_this.qitmeerdStatusAlert) {
            _this.qitmeerdStatusAlert.close();
            _this.qitmeerdStatusAlert = null;
          }

          if (typeof response.data.error != "undefined") {
            console.log(response.data.error);
            _this.qitmeerdStatusAlert = _this.$message({
              message: "获取节点信息异常! msg:" + response.data.error.message,
              type: "warning",
              duration: 0,
              onClose: function() {}
            });
            return;
          }

          _this.qitmeerdStatus = response.data.result;
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


