<template>
  <el-container>
    <el-header class="cheader">
      <el-row>
        <el-col :span="6">
          <h2>新建节点</h2>
        </el-col>
        <el-col :span="6"></el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名字">
          <el-input placeholder="nodename" v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item label="RPC地址">
          <el-input placeholder="http://127.0.0.1:18130" v-model="form.addr"></el-input>
        </el-form-item>
        <el-form-item label="RPC用户名">
          <el-input placeholder="RPC user" v-model="form.user"></el-input>
        </el-form-item>
        <el-form-item label="RPC密码">
          <el-input placeholder="RPC password" v-model="form.password"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary">创建</el-button>
          <el-button @click="nodeList">取消</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    return {
      form: {
        name: "",
        addr: "",
        user: "",
        password: ""
      }
    };
  },
  methods: {
    nodeList() {
      this.$router.push({ path: "/node" });
    }
  }
};
</script>