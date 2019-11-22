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
      <el-form :model="ruleForm" ref="ruleForm" :rules="rules" label-width="100px">
        <el-form-item label="助记词" prop="mnemonic">
          <el-input
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 4}"
            v-model="ruleForm.mnemonic"
          ></el-input>
        </el-form-item>
        <el-form-item label="请输入密码" prop="password1">
          <el-input placeholder="请输入密码" v-model="ruleForm.password1" show-password></el-input>
        </el-form-item>
        <el-form-item label="再次输入密码" prop="password2">
          <el-input placeholder="再次输入密码" v-model="ruleForm.password2" show-password></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submitForm('ruleForm')">恢复</el-button>
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
    let validatePass = (rule, value, callback) => {
      if (value === "") {
        callback(new Error("请输入密码"));
      } else {
        if (this.ruleForm.password2 !== "") {
          this.$refs.ruleForm.validateField("password2");
        }
        callback();
      }
    };
    let validatePass2 = (rule, value, callback) => {
      if (value === "") {
        callback(new Error("请再次输入密码"));
      } else if (value !== this.ruleForm.password1) {
        callback(new Error("两次输入密码不一致!"));
      } else {
        callback();
      }
    };

    return {
      ruleForm: {
        mnemonic: "",
        password1: "",
        password2: ""
      },
      rules: {
        password1: [{ validator: validatePass, trigger: "blur" }],
        password2: [{ validator: validatePass2, trigger: "blur" }]
      }
    };
  },
  mounted() {
    // this.$emit("checkWalletStats", lockStat => {
    //   this.$router.push("/");
    // });
  },
  methods: {
    submitForm(formName) {
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
            method: "wallet_recove",
            params: [this.ruleForm.mnemonic, this.ruleForm.password1]
          })
        }).then(response => {
          if (typeof response.data.error != "undefined") {
            this.$message({
              message: "错误，请稍后重试: " + response.data.error.message,
              type: "warning",
              duration: 500,
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
            duration: 500,
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