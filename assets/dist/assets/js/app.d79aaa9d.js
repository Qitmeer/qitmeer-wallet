(function(e){function t(t){for(var n,o,s=t[0],c=t[1],i=t[2],d=0,p=[];d<s.length;d++)o=s[d],l[o]&&p.push(l[o][0]),l[o]=0;for(n in c)Object.prototype.hasOwnProperty.call(c,n)&&(e[n]=c[n]);u&&u(t);while(p.length)p.shift()();return r.push.apply(r,i||[]),a()}function a(){for(var e,t=0;t<r.length;t++){for(var a=r[t],n=!0,s=1;s<a.length;s++){var c=a[s];0!==l[c]&&(n=!1)}n&&(r.splice(t--,1),e=o(o.s=a[0]))}return e}var n={},l={app:0},r=[];function o(t){if(n[t])return n[t].exports;var a=n[t]={i:t,l:!1,exports:{}};return e[t].call(a.exports,a,a.exports,o),a.l=!0,a.exports}o.m=e,o.c=n,o.d=function(e,t,a){o.o(e,t)||Object.defineProperty(e,t,{enumerable:!0,get:a})},o.r=function(e){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},o.t=function(e,t){if(1&t&&(e=o(e)),8&t)return e;if(4&t&&"object"===typeof e&&e&&e.__esModule)return e;var a=Object.create(null);if(o.r(a),Object.defineProperty(a,"default",{enumerable:!0,value:e}),2&t&&"string"!=typeof e)for(var n in e)o.d(a,n,function(t){return e[t]}.bind(null,n));return a},o.n=function(e){var t=e&&e.__esModule?function(){return e["default"]}:function(){return e};return o.d(t,"a",t),t},o.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},o.p="/app/";var s=window["webpackJsonp"]=window["webpackJsonp"]||[],c=s.push.bind(s);s.push=t,s=s.slice();for(var i=0;i<s.length;i++)t(s[i]);var u=c;r.push([0,"chunk-vendors"]),a()})({0:function(e,t,a){e.exports=a("56d7")},"56d7":function(e,t,a){"use strict";a.r(t);a("cadf"),a("551c"),a("f751"),a("097d");var n=a("2b0e"),l=a("5c96"),r=a.n(l),o=a("bc3a"),s=a.n(o),c=(a("0fae"),a("be35"),a("845fb"),function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{directives:[{name:"loading",rawName:"v-loading",value:e.loading,expression:"loading"}],attrs:{id:"app","element-loading-text":"加载钱包","element-loading-spinner":"el-icon-loading","element-loading-background":"rgba(0, 0, 0, 0.8)"}},[a("div",{attrs:{id:"sidebar"}},[a("div",{attrs:{id:"logo"}},[a("h1",[a("router-link",{attrs:{to:"/"}},[e._v("QITMEER WALLET")])],1)]),a("div",{attrs:{id:"menu"}},[a("h2",[e._v("用户")]),a("router-link",{attrs:{to:"/account"}},[a("i",{staticClass:"iconfont icon-Account"}),e._v(" \n        账户(key)管理\n      ")]),a("router-link",{attrs:{to:"/address",icon:"el-icon-tickets"}},[a("i",{staticClass:"iconfont icon-qianbao"}),e._v(" \n        地址管理\n      ")]),a("router-link",{attrs:{to:"/tx/send"}},[a("i",{staticClass:"iconfont icon-jiaoyi"}),e._v(" \n        发送交易\n      ")]),a("router-link",{attrs:{to:"/tx/list"}},[a("i",{staticClass:"iconfont icon-zhangdan"}),e._v(" \n        交易记录\n      ")]),a("router-link",{attrs:{to:"/backup"}},[a("i",{staticClass:"iconfont icon-baobeiguanli-baobeibeifen"}),e._v(" \n        备份/恢复\n      ")]),a("h2",[e._v("数据")]),a("router-link",{attrs:{to:"/node"}},[a("i",{staticClass:"iconfont icon-jiedian"}),e._v(" \n        节点管理\n      ")])],1),a("div",{attrs:{id:"info"}},[a("h2",[e._v("\n        节点: "+e._s(e.nodeName)+"\n        "),a("el-dropdown",[a("span",{staticClass:"el-dropdown-link"},[a("i",{staticClass:"el-icon-s-tools"})]),a("el-dropdown-menu",{attrs:{slot:"dropdown"},slot:"dropdown"},e._l(e.nodes,function(t,n){return a("el-dropdown-item",{key:n},[e._v(e._s(t))])}),1)],1)],1),e._m(0),e._m(1),e._m(2)]),e._m(3)]),a("router-view",{staticClass:"mainwarp"})],1)}),i=[function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("p",[a("span",[e._v("网络:")]),e._v("Mainnet\n      ")])},function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("p",[a("span",[e._v("全网:")]),e._v("100000\n      ")])},function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("p",[a("span",[e._v("节点:")]),e._v("1000\n      ")])},function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"version"},[a("p",[a("span",[e._v("版本:")]),e._v("201907-0.0.1\n      ")])])}],u={name:"app",data:function(){var e=["loacl","dao"];return{nodeName:"local",nodes:e,loading:!0,needOneAccount:!0}},mounted:function(){var e=this;this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_listAccount",params:null})}).then(function(t){e.loading=!1,"undefined"!=typeof t.data.error?e.$alert("获取账号信息异常，请刷新","账号",{showClose:!1,confirmButtonText:"确定",callback:function(t){e.$router.go(0)}}):0==t.data.result.length?e.$alert("创建至少一个账号","账号",{showClose:!1,confirmButtonText:"确定",callback:function(t){e.$router.push({path:"/account/new"}),e.$router.beforeEach(function(e,t,a){"/account/new"!=e.path&&a("/account/new")})}}):e.$store.state.accounts=t.data.result})},methods:{wathStats:function(){}},watch:{needOneAccount:function(){if(0==this.$store.state.accounts.length)return!1}}},d=u,p=a("2877"),m=Object(p["a"])(d,c,i,!1,null,null,null),f=m.exports,h=a("8c4f"),b=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("el-container",[n("el-header",{staticClass:"cheader"},[n("el-row",{attrs:{type:"flex"}},[n("el-col",{attrs:{span:6}},[n("h2",[e._v("首页")])]),n("el-col",{attrs:{span:12}}),n("el-col",{attrs:{span:6}})],1)],1),n("el-main",{staticClass:"cmain"},[n("div",[n("img",{staticStyle:{width:"80px"},attrs:{src:a("cf05")}}),n("h1",[e._v("Qitmeer Wallet")]),n("p",[e._v("欢迎使用Qitmeer Wallet")])])])],1)},v=[],w={data:function(){return{}},methods:{}},_=w,y=Object(p["a"])(_,b,v,!1,null,null,null),x=y.exports,g=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex"}},[a("el-col",{attrs:{span:6}},[a("h2",[e._v("账号管理")])]),a("el-col",{attrs:{span:12}}),a("el-col",{attrs:{span:6}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newAccount}},[e._v("新建账号")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.tableData}},[a("el-table-column",{attrs:{prop:"alias",label:"别名",width:"120"}}),a("el-table-column",{attrs:{prop:"key.xpub",label:"pubKey",width:"380"}}),a("el-table-column",{attrs:{prop:"balance",label:"余额"}})],1)],1)],1)},k=[],$={data:function(){var e=[];return{tableData:e}},methods:{newAccount:function(){this.$router.push({path:"/account/new"})}},mounted:function(){var e=this;this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_listAccount",params:null})}).then(function(t){"undefined"!=typeof t.data.error?e.$alert("错误，请稍后重试","seed",{showClose:!1,confirmButtonText:"确定",callback:function(e){}}):(e.tableData=t.data.result,e.$accounts=t.data.result)})}},C=$,F=Object(p["a"])(C,g,k,!1,null,null,null),E=F.exports,O=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",[a("el-col",{attrs:{span:6}},[a("h2",[e._v("新建账号")])]),a("el-col",{attrs:{span:6}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{ref:"ruleForm",attrs:{model:e.ruleForm,rules:e.rules,"label-width":"100px"}},[a("el-form-item",{attrs:{label:"随机种子",prop:"seed"}},[a("el-input",{attrs:{disabled:!0},model:{value:e.ruleForm.seed,callback:function(t){e.$set(e.ruleForm,"seed",t)},expression:"ruleForm.seed"}}),a("el-button",{attrs:{type:"primary",size:"small"},on:{click:e.newSeed}},[e._v("重新生成")])],1),a("el-form-item",{attrs:{label:"请输入密码",prop:"password1"}},[a("el-input",{attrs:{placeholder:"请输入密码","show-password":""},model:{value:e.ruleForm.password1,callback:function(t){e.$set(e.ruleForm,"password1",t)},expression:"ruleForm.password1"}})],1),a("el-form-item",{attrs:{label:"再次输入密码",prop:"password2"}},[a("el-input",{attrs:{placeholder:"再次输入密码","show-password":""},model:{value:e.ruleForm.password2,callback:function(t){e.$set(e.ruleForm,"password2",t)},expression:"ruleForm.password2"}})],1),a("el-form-item",{attrs:{label:"账号别名",prop:"alias"}},[a("el-input",{attrs:{placeholder:"12个字母数字以内"},model:{value:e.ruleForm.alias,callback:function(t){e.$set(e.ruleForm,"alias",t)},expression:"ruleForm.alias"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"},on:{click:function(t){return e.submitForm("ruleForm")}}},[e._v("创建")]),a("el-button",{on:{click:e.accountList}},[e._v("取消")])],1),a("div",[a("p",[e._v("注意：")]),a("p",[e._v("1. 账户=私钥（key)，是掌握您的账号的钥匙，请妥善保管。")]),a("p",[e._v("2. 私钥以加密形式保存，请在创建后，将私钥的加密文件单独备份。")]),a("p",[e._v("3. 密码用来加密您的私钥，请妥善设置，丢失无法找回。")]),a("p",[e._v("4. 别名仅在有多个账号的情况下，用做区分账号，仅此。")])])],1)],1)],1)},j=[],A={data:function(){var e=this,t=function(t,a,n){""===a?n(new Error("请输入密码")):(""!==e.ruleForm.password2&&e.$refs.ruleForm.validateField("password2"),n())},a=function(t,a,n){""===a?n(new Error("请再次输入密码")):a!==e.ruleForm.password1?n(new Error("两次输入密码不一致!")):n()},n=function(e,t,a){t.length>12?a(new Error("别名最长12个字符")):a()};return{ruleForm:{seed:"",password1:"",password2:"",alias:""},rules:{password1:[{validator:t,trigger:"blur"}],password2:[{validator:a,trigger:"blur"}],alias:[{validator:n,trigger:"blur"}]}}},mounted:function(){this.newSeed()},methods:{newSeed:function(){var e=this;this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_makeEntropy",params:null})}).then(function(t){"undefined"!=typeof t.data.error?e.$alert("错误，请稍后重试","seed",{showClose:!1,confirmButtonText:"确定",callback:function(e){}}):e.ruleForm.seed=t.data.result})},submitForm:function(e){var t=this;this.$refs.ruleForm.validate(function(e){if(!e)return!1;t.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_newAccount",params:[t.ruleForm.alias,t.ruleForm.seed,t.ruleForm.password1]})}).then(function(e){"undefined"!=typeof e.data.error?t.$alert("错误，请稍后重试","新建账户",{showClose:!1,confirmButtonText:"确定",callback:function(e){}}):t.$alert("\b创建账号成功","新建账户",{showClose:!1,confirmButtonText:"确定",callback:function(e){t.$router.push({path:"/account"})}})})})},accountList:function(){this.$router.push({path:"/account"})},newAccount:function(){this.$router.push({path:"/account/new"})}}},S=A,P=Object(p["a"])(S,O,j,!1,null,null,null),N=P.exports,T=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex",justify:"space-between"}},[a("el-col",{attrs:{span:4}},[a("h2",[e._v("地址管理")])]),a("el-col",{attrs:{span:8}},[a("el-row",[a("el-col",{attrs:{span:8}},[a("span",{staticStyle:{color:"#303133","font-weight":"600"}},[e._v("当前账户：")])]),a("el-col",{attrs:{span:14}},[a("el-select",{attrs:{placeholder:"账号",size:"small"},model:{value:e.current,callback:function(t){e.current=t},expression:"current"}},e._l(e.accounts,function(e){return a("el-option",{key:e.alias,attrs:{label:e.alias,value:e.alias}})}),1)],1)],1)],1),a("el-col",{attrs:{span:4}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newAddress}},[e._v("新建地址")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.addresses}},[a("el-table-column",{attrs:{prop:"index",label:"index",width:"60"}}),a("el-table-column",{attrs:{prop:"address",label:"地址"}})],1)],1)],1)},z=[],B={data:function(){return{accounts:this.$store.state.accounts,current:this.$store.state.accounts[0].alias,addresses:[]}},methods:{newAddress:function(){this.addresses.push({index:this.addresses.length,address:"1Nh7uHdvY6fNwtQtM1G5EZAFPLC33B59rB"})}}},L=B,R=Object(p["a"])(L,T,z,!1,null,null,null),D=R.exports,J=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex",justify:"space-between"}},[a("el-col",{attrs:{span:4}},[a("h2",[e._v("发送交易")])]),a("el-col",{attrs:{span:8}},[a("el-row",[a("el-col",{attrs:{span:8}},[a("span",{staticStyle:{color:"#303133","font-weight":"600"}},[e._v("当前账户：")])]),a("el-col",{attrs:{span:14}},[a("el-select",{attrs:{placeholder:"账号",size:"small"},model:{value:e.current,callback:function(t){e.current=t},expression:"current"}},e._l(e.accounts,function(e){return a("el-option",{key:e.value,attrs:{label:e.label,value:e.value}})}),1)],1)],1)],1),a("el-col",{attrs:{span:4}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{attrs:{model:e.form,"label-width":"120px"}},[a("el-form-item",{attrs:{label:"可用余额"}},[a("el-input",{attrs:{disabled:!0},model:{value:e.form.balance,callback:function(t){e.$set(e.form,"balance",t)},expression:"form.balance"}})],1),a("el-form-item",{attrs:{label:"转给(to)"}},[a("el-input",{attrs:{placeholder:"address"},model:{value:e.form.to,callback:function(t){e.$set(e.form,"to",t)},expression:"form.to"}})],1),a("el-form-item",{attrs:{label:"转账金额(amount)"}},[a("el-input",{attrs:{placeholder:"amount"},model:{value:e.form.value,callback:function(t){e.$set(e.form,"value",t)},expression:"form.value"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"}},[e._v("创建")]),a("el-button",{on:{click:e.accountList}},[e._v("取消")])],1)],1)],1)],1)},M=[],Q={data:function(){var e=[{value:"default",label:"default"},{value:"daodao",label:"daodao"}];return{accounts:e,current:e[0].value,form:{}}},methods:{accountList:function(){this.$router.push({path:"/account"})}}},U=Q,Y=Object(p["a"])(U,J,M,!1,null,null,null),q=Y.exports,H=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex",justify:"space-between"}},[a("el-col",{attrs:{span:4}},[a("h2",[e._v("交易记录")])]),a("el-col",{attrs:{span:8}},[a("el-row",[a("el-col",{attrs:{span:8}},[a("span",{staticStyle:{color:"#303133","font-weight":"600"}},[e._v("当前账户：")])]),a("el-col",{attrs:{span:14}},[a("el-select",{attrs:{placeholder:"账号",size:"small"},model:{value:e.current,callback:function(t){e.current=t},expression:"current"}},e._l(e.accounts,function(e){return a("el-option",{key:e.value,attrs:{label:e.label,value:e.value}})}),1)],1)],1)],1),a("el-col",{attrs:{span:4}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.txlist}},[a("el-table-column",{attrs:{prop:"date",label:"时间",width:"160"}}),a("el-table-column",{attrs:{prop:"type",label:"-",width:"40"}}),a("el-table-column",{attrs:{prop:"to",label:"to",width:"240"}}),a("el-table-column",{attrs:{prop:"amount",label:"金额"}})],1)],1)],1)},I=[],W={data:function(){var e=[{type:"in",from:"asdfasdfasdfasdf",to:"asdfasdfasdfasdf",amount:34234,date:"2019-07-23 23:32:23"},{type:"out",from:"asdfasdfasdfasdf",to:"asdfasdfasdfasdf",amount:34234,date:"2019-07-23 23:32:23"}],t=[{value:"default",label:"default"},{value:"daodao",label:"daodao"}];return{txlist:e,accounts:t,current:t[0].value}},methods:{newAddress:function(){this.addresses.push({index:this.addresses.length,address:"1Nh7uHdvY6fNwtQtM1G5EZAFPLC33B59rB"})}}},Z=W,G=Object(p["a"])(Z,H,I,!1,null,null,null),K=G.exports,V=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex"}},[a("el-col",{attrs:{span:6}},[a("h2",[e._v("备份恢复")])]),a("el-col",{attrs:{span:12}}),a("el-col",{attrs:{span:6}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newAccount}},[e._v("导入账号")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.tableData}},[a("el-table-column",{attrs:{prop:"alias",label:"别名",width:"120"}}),a("el-table-column",{attrs:{prop:"key",label:"key(encrypt)",width:"380"}}),a("el-table-column",[a("el-link",{attrs:{type:"primary",icon:"el-icon-download"}},[e._v("导出备份")]),e._v("  \n        "),a("el-link",{attrs:{type:"danger",icon:"el-icon-delete"}},[e._v("删除")])],1)],1)],1)],1)},X=[],ee={data:function(){var e=[{alias:"default",key:" tNNLc7iAQ5zzPeaHnz/UkqJrRezBSlpIzTMacSXRNhk=",balance:1231.12},{alias:"daodao",key:" 6evV6E7IOubFqBPF47N4TKoJcL4hUJYZlakYVgEi3Bo=",balance:123123123.12}];return{tableData:e}},methods:{newAccount:function(){this.$router.push({path:"/account/new"})}}},te=ee,ae=Object(p["a"])(te,V,X,!1,null,null,null),ne=ae.exports,le=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex"}},[a("el-col",{attrs:{span:6}},[a("h2",[e._v("节点管理")])]),a("el-col",{attrs:{span:12}}),a("el-col",{attrs:{span:6}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newNode}},[e._v("新建节点")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.tableData,"highlight-current-row":""},on:{"current-change":e.handleCurrentChange}},[a("el-table-column",{attrs:{type:"index",width:"50"}}),a("el-table-column",{attrs:{prop:"name",label:"名称",width:"120"}}),a("el-table-column",{attrs:{prop:"addr",label:"地址"}}),a("el-table-column",{attrs:{prop:"user",label:"user",width:"120"}}),a("el-table-column",{attrs:{prop:"pwd",label:"pwd",width:"120"}}),a("el-table-column",[a("el-link",{attrs:{type:"danger",icon:"el-icon-delete",width:"80"}},[e._v("删除")])],1)],1)],1)],1)},re=[],oe={data:function(){var e=[{name:"本地",addr:"127.0.0.1:1236",user:"admin",pwd:"123456"},{name:"daodao",addr:"12.4.23.4:1236",user:"admin",pwd:"123456"}];return{tableData:e}},methods:{newNode:function(){this.$router.push({path:"/node/new"})},handleCurrentChange:function(e){this.currentRow=e}}},se=oe,ce=Object(p["a"])(se,le,re,!1,null,null,null),ie=ce.exports,ue=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",[a("el-col",{attrs:{span:6}},[a("h2",[e._v("新建节点")])]),a("el-col",{attrs:{span:6}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{attrs:{model:e.form,"label-width":"100px"}},[a("el-form-item",{attrs:{label:"名字"}},[a("el-input",{attrs:{placeholder:"nodename"},model:{value:e.form.name,callback:function(t){e.$set(e.form,"name",t)},expression:"form.name"}})],1),a("el-form-item",{attrs:{label:"RPC地址"}},[a("el-input",{attrs:{placeholder:"http://127.0.0.1:18130"},model:{value:e.form.addr,callback:function(t){e.$set(e.form,"addr",t)},expression:"form.addr"}})],1),a("el-form-item",{attrs:{label:"RPC用户名"}},[a("el-input",{attrs:{placeholder:"RPC user"},model:{value:e.form.user,callback:function(t){e.$set(e.form,"user",t)},expression:"form.user"}})],1),a("el-form-item",{attrs:{label:"RPC密码"}},[a("el-input",{attrs:{placeholder:"RPC password"},model:{value:e.form.password,callback:function(t){e.$set(e.form,"password",t)},expression:"form.password"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"}},[e._v("创建")]),a("el-button",{on:{click:e.nodeList}},[e._v("取消")])],1)],1)],1)],1)},de=[],pe={data:function(){return{form:{name:"",addr:"",user:"",password:""}}},methods:{nodeList:function(){this.$router.push({path:"/node"})}}},me=pe,fe=Object(p["a"])(me,ue,de,!1,null,null,null),he=fe.exports;n["default"].use(h["a"]);var be=new h["a"]({routes:[{path:"/",name:"index",component:x},{path:"/account",name:"account",component:E},{path:"/account/new",name:"accountnew",component:N},{path:"/address",name:"address",component:D},{path:"/tx/send",name:"txsend",component:q},{path:"/tx/list",name:"txlist",component:K},{path:"/backup",name:"backup",component:ne},{path:"/node",name:"node",component:ie},{path:"/node/new",name:"nodenew",component:he}]}),ve=a("2f62");n["default"].use(ve["a"]);var we={Accounts:[]},_e=new ve["a"].Store({state:we}),ye=_e;n["default"].use(r.a),n["default"].prototype.$axios=s.a.create({baseURL:window.QitmeerConfig.RPCAddr,timeout:1e3,headers:{"content-type":"application/json"},auth:{username:window.QitmeerConfig.RPCUser,password:window.QitmeerConfig.RPCPass},withCredentials:!0,crossDomain:!0}),new n["default"]({el:"#app",router:be,store:ye,render:function(e){return e(f)}})},"845fb":function(e,t,a){},be35:function(e,t,a){},cf05:function(e,t,a){e.exports=a.p+"assets/img/logo.c9da5caa.png"}});
//# sourceMappingURL=app.d79aaa9d.js.map