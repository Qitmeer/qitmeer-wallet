<template>
  <el-container>
    <el-header class="cheader">
      <el-row>
        <el-col :span="6">
          <h2>新建钱包</h2>
        </el-col>
        <el-col :span="6"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-form :model="ruleForm" ref="ruleForm" :rules="rules" label-width="100px">
        <el-form-item label="钱包种子" prop="seed">
          <el-input v-model="ruleForm.seed" :disabled="true"></el-input>
          <el-button type="primary" size="small" @click="newSeed">重新生成</el-button>
        </el-form-item>
        <el-form-item label="助记词" prop="mnemonic">
          <el-input
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 4}"
            v-model="ruleForm.mnemonic"
            :readonly="true"
          ></el-input>
        </el-form-item>
        <el-form-item label="请输入密码" prop="password1">
          <el-input placeholder="请输入密码" v-model="ruleForm.password1" show-password></el-input>
        </el-form-item>
        <el-form-item label="再次输入密码" prop="password2">
          <el-input placeholder="再次输入密码" v-model="ruleForm.password2" show-password></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submitForm('ruleForm')">创建</el-button>
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
    var validatePass = (rule, value, callback) => {
      if (value === "") {
        callback(new Error("请输入密码"));
      } else {
        if (this.ruleForm.password2 !== "") {
          this.$refs.ruleForm.validateField("password2");
        }
        callback();
      }
    };
    var validatePass2 = (rule, value, callback) => {
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
        seed: "",
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
    // this.$emit("checkWalletStats");
    this.newSeed();
  },
  methods: {
    newSeed() {
      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "wallet_makeSeed",
          params: null
        })
      }).then(response => {
        if (typeof response.data.error != "undefined") {
          this.$alert("错误，请稍后重试", "seed", {
            showClose: false,
            confirmButtonText: "确定",
            callback: action => {}
          });
        } else {
          this.ruleForm.seed = response.data.result.seed;
          this.ruleForm.mnemonic = response.data.result.mnemonic;
        }
      });
    },
    submitForm(formName) {
      let _this = this;
      this.$refs.ruleForm.validate(valid => {
        if (!valid) {
          return false;
        }

        this.$emit("setLoading", true, "创建钱包");

        this.$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "wallet_create",
            params: [this.ruleForm.seed, this.ruleForm.password1]
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
            message: "创建成功!",
            type: "success",
            duration: 500,
            onClose: function() {
              _this.$emit("setLoading", false, "");
              _this.$emit("getWalletStats");
            }
          });
        });
      });
    }
  }
};
</script>