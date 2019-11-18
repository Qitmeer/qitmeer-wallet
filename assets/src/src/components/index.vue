<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex">
        <el-col :span="6">
          <h2>首页</h2>
        </el-col>
        <el-col :span="12"></el-col>
        <el-col :span="6"></el-col>
      </el-row>
    </el-header>
    <el-main class="cmain">
      <div>
        <img src="../assets/logo.png" style="width:80px;" />
        <h1>Qitmeer Wallet</h1>
        <p>欢迎使用Qitmeer Wallet</p>
      </div>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    return {};
  },
  mounted() {
    this.$emit("checkWalletStats", lockStat => {
      console.log("index: wallet opened");
    });
  },
  methods: {}
};
</script>