<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex">
        <el-col :span="6">
          <h2>节点管理</h2>
        </el-col>
        <el-col :span="12"></el-col>
        <el-col :span="6">
          <el-button type="primary" icon="el-icon-plus" @click="newNode" size="small">新建节点</el-button>
        </el-col>
      </el-row>
    </el-header>
    <el-main class="cmain">
      <el-table :data="tableData" highlight-current-row @current-change="handleCurrentChange">
        <el-table-column type="index" width="50"></el-table-column>
        <el-table-column prop="name" label="名称" width="120"></el-table-column>
        <el-table-column prop="addr" label="地址"></el-table-column>
        <el-table-column prop="user" label="user" width="120"></el-table-column>
        <el-table-column prop="pwd" label="pwd" width="120"></el-table-column>
        <el-table-column>
          <el-link type="danger" icon="el-icon-delete" width="80">删除</el-link>
        </el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>
<style>
</style>

<script>
export default {
  data() {
    const item = [
      {
        name: "本地",
        addr: "127.0.0.1:1236",
        user: "admin",
        pwd: "123456"
      },
      {
        name: "daodao",
        addr: "12.4.23.4:1236",
        user: "admin",
        pwd: "123456"
      }
    ];
    return {
      tableData: item
    };
  },
  methods: {
    newNode() {
      this.$router.push({ path: "/node/new" });
    },
    handleCurrentChange(val) {
      this.currentRow = val;
    }
  }
};
</script>