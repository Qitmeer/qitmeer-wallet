<template>
  <el-container>
    <el-header class="cheader">
      <el-row>
        <el-col :span="6">
          <h2>导入私钥</h2>
        </el-col>
        <el-col :span="6"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-form :model="ruleForm" ref="ruleForm" :rules="rules" label-width="100px">
        <el-form-item label="私钥" prop="key">
          <el-input type="textarea" :autosize="{ minRows: 4, maxRows: 6}" v-model="ruleForm.key"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submitForm">导入</el-button>
          <el-button @click="goBackup">取消</el-button>
        </el-form-item>
        <div>
          <p>注意:</p>
          <p>1. 目前只支持导入地址到imported账户</p>
        </div>
      </el-form>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    var checkAccount = (rule, value, callback) => {
      // if (value == "*") {
      //   callback(new Error("*"));
      // }
      callback();
    };
    return {
      ruleForm: {
        keys: ""
      },
      rules: {
        keys: [{ validator: checkAccount, trigger: "blur" }]
      }
    };
  },
  mounted() {
    let _this = this;
    _this.$emit("getWalletStats", stats => {
      if (stats != "unlock") {
        _this.$emit("walletPasswordDlg", "wallet_unlock", result => {
          if (!result) {
            _this.$router.push("/backup");
          }
        });
      }
    });
  },
  methods: {
    submitForm() {
      let _this = this;
      this.$refs.ruleForm.validate(valid => {
        if (!valid) {
          return false;
        }

        this.$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "wallet_importWifPrivKey",
            params: ["imported", this.ruleForm.key]
          })
        }).then(response => {
          if (typeof response.data.error != "undefined") {
            _this.$emit("alertResError", response.data.error, () => {
              _this.$router.push("/backup");
            });
            return;
          }
          this.$alert("导入私钥成功", {
            showClose: false,
            closeOnClickModal: false,
            closeOnPressEscape: false,
            confirmButtonText: "确定",
            callback: (action, instance) => {
              this.$router.push({ path: "/backup" });
            }
          });
        });
      });
    },
    goBackup() {
      this.$router.push({ path: "/backup" });
    }
  }
};
</script>