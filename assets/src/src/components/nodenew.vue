<template>
  <el-container>
    <el-header class="cheader">
      <el-row>
        <el-col :span="6">
          <h2>{{act.title}}节点</h2>
        </el-col>
        <el-col :span="6"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名字">
          <el-input v-model="form.Name" :disabled="act.nameDisable"></el-input>
        </el-form-item>
        <el-form-item label="RPC地址">
          <el-input placeholder="127.0.0.1:18130" v-model="form.RPCServer"></el-input>
        </el-form-item>
        <el-form-item label="RPC用户名">
          <el-input placeholder="RPC user" v-model="form.RPCUser"></el-input>
        </el-form-item>
        <el-form-item label="RPC密码">
          <el-input placeholder="RPC password" v-model="form.RPCPassword"></el-input>
        </el-form-item>

        <el-form-item label="TLS">
          <el-checkbox v-model="form.NoTLS">NoTLS</el-checkbox>
          <el-checkbox v-model="form.TLSSkipVerify">TLSSkipVerify</el-checkbox>
        </el-form-item>

        <el-form-item label="Proxy">
          <el-input v-model="form.Proxy"></el-input>
        </el-form-item>
        <el-form-item label="ProxyUser">
          <el-input v-model="form.ProxyUser"></el-input>
        </el-form-item>
        <el-form-item label="ProxyPass">
          <el-input v-model="form.ProxyPass"></el-input>
        </el-form-item>

        <el-form-item>
          <el-button @click="submit" type="primary">{{act.title}}</el-button>
          <el-button @click="nodeList">取消</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
let action = {
  create: {
    nameDisable: false,
    method: "qitmeerd_add",
    title: "添加"
  },
  update: {
    nameDisable: true,
    method: "qitmeerd_update",
    title: "更新"
  }
};
export default {
  data() {
    return {
      act: action.create,
      form: {
        Name: "",
        RPCServer: "",
        RPCUser: "",
        RPCPassword: "",
        RPCCert: "",
        NoTLS: true,
        TLSSkipVerify: true,
        Proxy: "",
        ProxyUser: "",
        ProxyPass: ""
      }
    };
  },
  mounted() {
    if (this.$route.params.name) {
      if (this.$store.state.QitmeerdList[this.$route.params.name]) {
        this.act = action.update;
        let item = this.$store.state.QitmeerdList[this.$route.params.name];
        this.form = {
          Name: this.$route.params.name,
          RPCServer: item.RPCServer,
          RPCUser: item.RPCUser,
          RPCPassword: item.RPCPassword,
          RPCCert: item.RPCCert,
          NoTLS: item.NoTLS,
          TLSSkipVerify: item.TLSSkipVerify,
          Proxy: item.Proxy,
          ProxyUser: item.ProxyUser,
          ProxyPass: item.ProxyPass
        };
      }
    }
  },
  methods: {
    nodeList() {
      this.$router.push({ path: "/node" });
    },
    submit() {
      let _this = this;

      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: _this.act.method,
          params: [
            this.form.Name,
            this.form.RPCServer,
            this.form.RPCUser,
            this.form.RPCPassword,
            this.form.RPCCert,
            this.form.NoTLS,
            this.form.TLSSkipVerify,
            this.form.Proxy,
            this.form.ProxyUser,
            this.form.ProxyPass
          ]
        })
      }).then(response => {
        if (typeof response.data.error != "undefined") {
          _this.$emit("alertResError", response.data.error, () => {
            _this.$router.push("/node");
          });
          return;
        }
        this.$alert("成功", {
          showClose: false,
          closeOnClickModal: false,
          closeOnPressEscape: false,
          confirmButtonText: "确定",
          callback: (action, instance) => {
            this.$router.push({ path: "/node" });
          }
        });
      });
    }
  }
};
</script>