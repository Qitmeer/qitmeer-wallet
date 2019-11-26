<template>
  <el-container>
    <el-header class="cheader">
      <el-row type="flex" justify="space-between">
        <el-col :span="4">
          <h2>地址管理</h2>
        </el-col>
        <el-col :span="8">
          <el-row>
            <el-col :span="8">
              <span style="color:#303133;font-weight:600;">当前账户：</span>
            </el-col>
            <el-col :span="14">
              <el-select v-model="current" placeholder="账号" size="small" @change="getAddressList">
                <el-option
                  v-for="item in accounts"
                  :key="item.index"
                  :label="item.account"
                  :value="item.account"
                ></el-option>
              </el-select>
            </el-col>
          </el-row>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" icon="el-icon-plus" @click="newAddress" size="small">新建地址</el-button>
        </el-col>
      </el-row>
    </el-header>

    <el-main class="cmain">
      <el-table :data="addresses">
        <el-table-column type="index" width="50"></el-table-column>
        <el-table-column prop="addr" label="地址"></el-table-column>
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
      accounts: [],
      current: "",
      addresses: []
    };
  },
  mounted() {
    if (this.$store.state.Accounts.length == 0) {
      this.$emit("updateAccounts", () => {
        this.accounts = this.$store.state.Accounts;
        this.current = this.$store.state.Accounts[0].account;

        this.getAddressList();
      });
    } else {
      this.accounts = this.$store.state.Accounts;
      this.current = this.$store.state.Accounts[0].account;

      this.getAddressList();
    }
  },
  methods: {
    newAddress() {
      let _this = this;
      this.$axios({
        method: "post",
        data: JSON.stringify({
          id: new Date().getTime(),
          method: "wallet_createAddress",
          params: [_this.current]
        })
      }).then(response => {
        if (typeof response.data.error != "undefined") {
          _this.$emit("alertResError", response.data.error, () => {});
          return;
        }
        this.addresses.push({
          addr: response.data.result
        });
      });
    },
    getAddressList() {
      let _this = this;
      _this
        .$axios({
          method: "post",
          data: JSON.stringify({
            id: new Date().getTime(),
            method: "wallet_getAddressesByAccount",
            params: [_this.current]
          })
        })
        .then(response => {
          if (typeof response.data.error != "undefined") {
            _this.$emit("alertResError", response.data.error, () => {});
            return;
          }

          let tmpTable = [];
          for (let i = 0; i < response.data.result.length; i++) {
            tmpTable.push({ addr: response.data.result[i] });
          }

          _this.addresses = tmpTable;
        });
    }
  }
};
</script>