<template>
  <el-container>
    <el-header class="cheader">
      <el-row>
        <el-col :span="6">
          <h2>新建账号</h2>
        </el-col>
        <el-col :span="6"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-form :model="ruleForm" ref="ruleForm" :rules="rules" label-width="100px">
        <el-form-item label="账号名称" prop="name">
          <el-input placeholder="账号" v-model="ruleForm.name"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submitForm('ruleForm')">创建</el-button>
          <el-button @click="accountList">取消</el-button>
        </el-form-item>
        <div></div>
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
      if (value == "*") {
        callback(new Error("账号名不能为*"));
      }
      callback();
    };
    return {
      ruleForm: {
        account: ""
      },
      rules: {
        account: [{ validator: checkAccount, trigger: "blur" }]
      }
    };
  },
  mounted() {
    let _this = this;
    _this.$emit("getWalletStats", stats => {
      if (stats != "unlock") {
        _this.$emit("walletPasswordDlg", "wallet_unlock", result => {
          if (!result) {
            _this.$router.push("/account");
          }
        });
      }
    });
  },
  methods: {
    submitForm(formName) {
      let _this = this;
      this.$refs.ruleForm.validate(valid => {
        if (!valid) {
          return false;
        }

        this.$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "wallet_create",
            params: [this.ruleForm.name]
          })
        }).then(response => {
          if (typeof response.data.error != "undefined") {
            _this.$emit("alertResError", response.data.error, () => {
              _this.$router.push("/account");
            });
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
      });
    },
    accountList() {
      this.$router.push({ path: "/account" });
    },
    newAccount() {
      this.$router.push({ path: "/account/new" });
    }
  }
};
</script>