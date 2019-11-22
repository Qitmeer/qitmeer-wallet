<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex">
        <el-col :span="6">
          <h2>节点管理</h2>
        </el-col>
        <el-col :span="12"></el-col>
        <el-col :span="6">
          <el-button type="primary" icon="el-icon-plus" @click="newNode" size="small">添加节点</el-button>
        </el-col>
      </el-row>
    </el-header>
    <el-main class="cmain">
      <el-table :data="tableData" style="width: 100%">
        <el-table-column type="index" width="40"></el-table-column>
        <el-table-column prop="Name" label="名称" width="80"></el-table-column>
        <el-table-column prop="RPCServer" label="地址" width="200"></el-table-column>
        <el-table-column prop="RPCUser" label="user" width="120"></el-table-column>
        <el-table-column prop="RPCPassword" label="pwd" width="120"></el-table-column>
        <el-table-column label="操作" width="100">
          <template slot-scope="scope">
            <el-button type="text" @click="delNode(scope.row)">删除</el-button>
            <el-button type="text" @click="editNode(scope.row)">编辑</el-button>
          </template>
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
    return {
      tableData: []
    };
  },
  mounted() {
    let _this = this;
    _this.$emit("getQitmeerdList", () => {
      _this.tableData = _this.$store.state.QitmeerdList;
    });
  },
  methods: {
    newNode() {
      this.$router.push({ path: "/node/new" });
    },
    editNode(node) {
      this.$router.push({
        path: `/node/edit/${node.Name}`
      });
    },
    delNode(node) {
      let _this = this;

      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "qitmeerd_del",
          params: [node.Name]
        })
      }).then(response => {
        if (typeof response.data.error != "undefined") {
          _this.$emit("Del error", response.data.error, () => {});
          return;
        }
        this.$alert("成功", {
          showClose: false,
          closeOnClickModal: false,
          closeOnPressEscape: false,
          confirmButtonText: "确定",
          callback: (action, instance) => {
            _this.$emit("getQitmeerdList", () => {
              _this.tableData = _this.$store.state.QitmeerdList;
            });
          }
        });
      });
    }
  }
};
</script>