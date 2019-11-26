<template>
  <el-container>
    <el-header class="cheader">
      <el-row>
        <el-col :span="6">
          <h2>恢复钱包</h2>
        </el-col>
        <el-col :span="6"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-form :model="ruleForm" ref="ruleForm" :rules="rules" label-width="150px">
        <el-form-item label="助记词" prop="mnemonic">
          <el-input
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 4}"
            v-model="ruleForm.mnemonic"
          ></el-input>
        </el-form-item>
        <el-form-item label="登录密码" prop="password1">
          <el-input placeholder="登录密码" v-model="ruleForm.password1" show-password></el-input>
        </el-form-item>
        <el-form-item label="再次输入登录密码" prop="password2">
          <el-input placeholder="登录密码" v-model="ruleForm.password2" show-password></el-input>
        </el-form-item>
        <el-form-item label="交易密码" prop="password21">
          <el-input placeholder="交易密码" v-model="ruleForm.password21" show-password></el-input>
        </el-form-item>
        <el-form-item label="再次输入交易密码" prop="password22">
          <el-input placeholder="交易密码" v-model="ruleForm.password22" show-password></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submitForm">恢复</el-button>
        </el-form-item>
        <div>
          <p>注意：</p>
          <p>1. 助记词用来备份恢复钱包，请妥善安全保管。</p>
          <p>2. 密码只用来加密您的本地钱包数据。</p>
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
    var validatePass = p2 => {
      return (rule, value, callback) => {
        if (value === "") {
          callback(new Error("请输入密码"));
        } else {
          if (this.ruleForm[p2] !== "") {
            this.$refs.ruleForm.validateField(p2);
          }
          callback();
        }
      };
    };
    var validatePass2 = p1 => {
      return (rule, value, callback) => {
        if (value === "") {
          callback(new Error("请再次输入密码"));
        } else if (value !== this.ruleForm[p1]) {
          callback(new Error("两次输入密码不一致!"));
        } else {
          callback();
        }
      };
    };

    return {
      ruleForm: {
        mnemonic: "",
        password1: "",
        password2: "",
        password21: "",
        password22: ""
      },
      rules: {
        password1: [{ validator: validatePass("password2"), trigger: "blur" }],
        password2: [{ validator: validatePass2("password1"), trigger: "blur" }],
        password21: [
          { validator: validatePass("password22"), trigger: "blur" }
        ],
        password22: [
          { validator: validatePass2("password21"), trigger: "blur" }
        ]
      }
    };
  },
  mounted() {},
  methods: {
    submitForm() {
      let _this = this;
      this.$refs.ruleForm.validate(valid => {
        if (!valid) {
          return false;
        }

        this.$emit("setLoading", true, "恢复钱包");

        this.$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "ui_recoverWallet",
            params: [
              this.ruleForm.mnemonic,
              this.ruleForm.password1,
              this.ruleForm.password21
            ]
          })
        }).then(response => {
          if (typeof response.data.error != "undefined") {
            this.$message({
              message: "错误，请稍后重试: " + response.data.error.message,
              type: "warning",
              duration: 1000,
              onClose: function() {
                _this.$emit("setLoading", false);
                _this.$router.go(0);
              }
            });
            return;
          }
          this.$message({
            message: "恢复成功成功!",
            type: "success",
            duration: 1000,
            onClose: function() {
              _this.$emit("setLoading", false, "");
              _this.$emit("getWalletStats", action => {
                _this.$router.push("/");
              });
            }
          });
        });
      });
    }
  }
};
</script>