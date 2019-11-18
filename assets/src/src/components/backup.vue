<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex">
        <el-col :span="6">
          <h2>备份恢复</h2>
        </el-col>
        <el-col :span="12"></el-col>
        <el-col :span="6">
          <el-button type="primary" icon="el-icon-plus" @click="newAccount" size="small">导入账号</el-button>
        </el-col>
      </el-row>
    </el-header>
    <el-main class="cmain">
      <el-table :data="tableData">
        <el-table-column prop="alias" label="别名" width="120"></el-table-column>
        <el-table-column prop="key" label="key(encrypt)" width="380"></el-table-column>
        <el-table-column>
          <el-link type="primary" icon="el-icon-download">导出备份</el-link>&nbsp;&nbsp;
          <el-link type="danger" icon="el-icon-delete">删除</el-link>
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
        alias: "default",
        key: " tNNLc7iAQ5zzPeaHnz/UkqJrRezBSlpIzTMacSXRNhk=",
        balance: 1231.12
      },
      {
        alias: "daodao",
        key: " 6evV6E7IOubFqBPF47N4TKoJcL4hUJYZlakYVgEi3Bo=",
        balance: 123123123.12
      }
    ];
    return {
      tableData: item
    };
  },
  methods: {
    newAccount() {
      this.$router.push({ path: "/account/new" });
    }
  }
};
</script>