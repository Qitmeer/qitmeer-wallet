(function(e){function t(t){for(var n,l,s=t[0],c=t[1],i=t[2],d=0,m=[];d<s.length;d++)l=s[d],r[l]&&m.push(r[l][0]),r[l]=0;for(n in c)Object.prototype.hasOwnProperty.call(c,n)&&(e[n]=c[n]);u&&u(t);while(m.length)m.shift()();return o.push.apply(o,i||[]),a()}function a(){for(var e,t=0;t<o.length;t++){for(var a=o[t],n=!0,s=1;s<a.length;s++){var c=a[s];0!==r[c]&&(n=!1)}n&&(o.splice(t--,1),e=l(l.s=a[0]))}return e}var n={},r={app:0},o=[];function l(t){if(n[t])return n[t].exports;var a=n[t]={i:t,l:!1,exports:{}};return e[t].call(a.exports,a,a.exports,l),a.l=!0,a.exports}l.m=e,l.c=n,l.d=function(e,t,a){l.o(e,t)||Object.defineProperty(e,t,{enumerable:!0,get:a})},l.r=function(e){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(e,"__esModule",{value:!0})},l.t=function(e,t){if(1&t&&(e=l(e)),8&t)return e;if(4&t&&"object"===typeof e&&e&&e.__esModule)return e;var a=Object.create(null);if(l.r(a),Object.defineProperty(a,"default",{enumerable:!0,value:e}),2&t&&"string"!=typeof e)for(var n in e)l.d(a,n,function(t){return e[t]}.bind(null,n));return a},l.n=function(e){var t=e&&e.__esModule?function(){return e["default"]}:function(){return e};return l.d(t,"a",t),t},l.o=function(e,t){return Object.prototype.hasOwnProperty.call(e,t)},l.p="/app/";var s=window["webpackJsonp"]=window["webpackJsonp"]||[],c=s.push.bind(s);s.push=t,s=s.slice();for(var i=0;i<s.length;i++)t(s[i]);var u=c;o.push([0,"chunk-vendors"]),a()})({0:function(e,t,a){e.exports=a("56d7")},"56d7":function(e,t,a){"use strict";a.r(t);a("cadf"),a("551c"),a("f751"),a("097d");var n=a("2b0e"),r=a("5c96"),o=a.n(r),l=a("bc3a"),s=a.n(l),c=(a("0fae"),a("be35"),a("845fb"),function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{directives:[{name:"loading",rawName:"v-loading",value:e.loading,expression:"loading"}],attrs:{id:"app","element-loading-text":e.loadingText,"element-loading-spinner":"el-icon-loading","element-loading-background":"rgba(0, 0, 0, 0.8)"}},[a("div",{attrs:{id:"sidebar"}},[a("div",{attrs:{id:"logo"}},[a("h1",[a("router-link",{attrs:{to:"/"}},[e._v("QITMEER WALLET")])],1)]),a("div",{attrs:{id:"menu"}},[a("h2",[e._v("用户")]),a("router-link",{attrs:{to:"/account"}},[a("i",{staticClass:"iconfont icon-Account"}),e._v(" \n        账户管理\n      ")]),a("router-link",{attrs:{to:"/address",icon:"el-icon-tickets"}},[a("i",{staticClass:"iconfont icon-qianbao"}),e._v(" \n        地址管理\n      ")]),a("router-link",{attrs:{to:"/tx/send"}},[a("i",{staticClass:"iconfont icon-jiaoyi"}),e._v(" \n        发送交易\n      ")]),a("router-link",{attrs:{to:"/tx/list"}},[a("i",{staticClass:"iconfont icon-zhangdan"}),e._v(" \n        交易记录\n      ")]),a("router-link",{attrs:{to:"/backup"}},[a("i",{staticClass:"iconfont icon-baobeiguanli-baobeibeifen"}),e._v(" \n        备份/恢复\n      ")]),a("h2",[e._v("数据")]),a("router-link",{attrs:{to:"/node"}},[a("i",{staticClass:"iconfont icon-jiedian"}),e._v(" \n        节点管理\n      ")])],1),a("div",{attrs:{id:"info"}},[a("h2",[e._v("\n        节点: "+e._s(e.nodeName)+"\n        "),a("el-dropdown",[a("span",{staticClass:"el-dropdown-link"},[a("i",{staticClass:"el-icon-s-tools"})]),a("el-dropdown-menu",{attrs:{slot:"dropdown"},slot:"dropdown"},e._l(e.nodes,function(t,n){return a("el-dropdown-item",{key:n},[e._v(e._s(t))])}),1)],1)],1),e._m(0),e._m(1),e._m(2)]),e._m(3)]),a("router-view",{staticClass:"mainwarp",on:{setLoading:e.setLoading,checkWalletStats:e.checkWalletStats,dlgUnlockWallet:e.dlgUnlockWallet,alertResError:e.alertResError}})],1)}),i=[function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("p",[a("span",[e._v("网络:")]),e._v("Mainnet\n      ")])},function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("p",[a("span",[e._v("全网:")]),e._v("100000\n      ")])},function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("p",[a("span",[e._v("节点:")]),e._v("1000\n      ")])},function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"version"},[a("p",[a("span",[e._v("版本:")]),e._v("201907-0.0.1\n      ")])])}],u={name:"app",data:function(){var e=["loacl","dao"];return{nodeName:"local",nodes:e,loading:!1,loadingText:"",walletcrate:!0,openForm:{walletPass:""}}},mounted:function(){},methods:{setLoading:function(e,t){this.loading=e,this.loadingText=t},wathStats:function(){},checkWalletStats:function(e){var t=this;t.loading=!0,t.loadingText="检查钱包状态",t.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"wallet_status",params:null})}).then(function(a){if("undefined"==typeof a.data.error){t.loading=!1,t.loadingText="";var n=a.data.result;switch(n.stats){case"nil":"/"==t.$route.path&&t.dlgCreateWallet(),"/wallet/create"!=t.$route.path&&"/wallet/recove"!=t.$route.path&&t.$router.push("/");break;case"closed":t.openWallet();break;case"lock":e("lock");break;case"unlock":e("unlock");break}}else t.$message({message:"获取钱包信息异常，请刷新! /n "+a.data.error,type:"warning",duration:500,onClose:function(){t.$router.push("/")}})})},openWallet:function(){var e=this,t=this;this.loading=!0,this.loadingText="打开钱包",this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"wallet_open",params:[this.openForm.walletPass]})}).then(function(a){e.loading=!1,"undefined"==typeof a.data.error?(console.log("open wallet ok!"),e.$router.push("/")):e.$message({message:"打开钱包错误："+a.data.error.message,type:"warning",duration:500,onClose:function(){t.$emit("setLoading",!1),t.$router.go(0)}})})},dlgCreateWallet:function(){var e=this;this.$confirm("","创建新钱包",{showClose:!1,closeOnClickModal:!1,confirmButtonText:"创建新钱包",cancelButtonText:"恢复钱包",callback:function(t,a){"cancel"==t?e.$router.push({path:"/wallet/recove"}):e.$router.push({path:"/wallet/create"})}})},dlgUnlockWallet:function(e){var t=this,a=this;a.$prompt("解锁钱包",{confirmButtonText:"确定",cancelButtonText:"取消"}).then(function(a){var n=a.value;t.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"wallet_unlock",params:[n,2e3]})}).then(function(a){if("undefined"!=typeof a.data.error)return t.$message({message:h("div",null,[h("p",null,"错误！"),h("p",null,"code:"+a.data.error.code),h("p",null,"info:"+a.data.error.message)]),type:"warning",duration:500,onClose:function(){}}),void e(!1);e(!0)})}).catch(function(){e(!1)})},alertResError:function(e,t){var a=this.$createElement;this.$alert(a("div",null,[a("p",null,"错误！"),a("p",null,"code:"+e.code),a("p",null,"info:"+e.message)]),{showClose:!1,closeOnClickModal:!1,closeOnPressEscape:!1,confirmButtonText:"确定",callback:t})}},watch:{needOneAccount:function(){if(0==this.$store.state.accounts.length)return!1}}},d=u,m=a("2877"),p=Object(m["a"])(d,c,i,!1,null,null,null),f=p.exports,w=a("8c4f"),b=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("el-container",[n("el-header",{staticClass:"cheader"},[n("el-row",{attrs:{type:"flex"}},[n("el-col",{attrs:{span:6}},[n("h2",[e._v("首页")])]),n("el-col",{attrs:{span:12}}),n("el-col",{attrs:{span:6}})],1)],1),n("el-main",{staticClass:"cmain"},[n("div",[n("img",{staticStyle:{width:"80px"},attrs:{src:a("cf05")}}),n("h1",[e._v("Qitmeer Wallet")]),n("p",[e._v("欢迎使用Qitmeer Wallet")])])])],1)},v=[],g={data:function(){return{}},mounted:function(){this.$emit("checkWalletStats",function(e){console.log("index: wallet opened")})},methods:{}},$=g,_=Object(m["a"])($,b,v,!1,null,null,null),x=_.exports,y=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",[a("el-col",{attrs:{span:6}},[a("h2",[e._v("新建钱包")])]),a("el-col",{attrs:{span:6}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{ref:"ruleForm",attrs:{model:e.ruleForm,rules:e.rules,"label-width":"100px"}},[a("el-form-item",{attrs:{label:"钱包种子",prop:"seed"}},[a("el-input",{attrs:{disabled:!0},model:{value:e.ruleForm.seed,callback:function(t){e.$set(e.ruleForm,"seed",t)},expression:"ruleForm.seed"}}),a("el-button",{attrs:{type:"primary",size:"small"},on:{click:e.newSeed}},[e._v("重新生成")])],1),a("el-form-item",{attrs:{label:"助记词",prop:"mnemonic"}},[a("el-input",{attrs:{type:"textarea",autosize:{minRows:2,maxRows:4},readonly:!0},model:{value:e.ruleForm.mnemonic,callback:function(t){e.$set(e.ruleForm,"mnemonic",t)},expression:"ruleForm.mnemonic"}})],1),a("el-form-item",{attrs:{label:"请输入密码",prop:"password1"}},[a("el-input",{attrs:{placeholder:"请输入密码","show-password":""},model:{value:e.ruleForm.password1,callback:function(t){e.$set(e.ruleForm,"password1",t)},expression:"ruleForm.password1"}})],1),a("el-form-item",{attrs:{label:"再次输入密码",prop:"password2"}},[a("el-input",{attrs:{placeholder:"再次输入密码","show-password":""},model:{value:e.ruleForm.password2,callback:function(t){e.$set(e.ruleForm,"password2",t)},expression:"ruleForm.password2"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"},on:{click:function(t){return e.submitForm("ruleForm")}}},[e._v("创建")])],1),a("div",[a("p",[e._v("注意：")]),a("p",[e._v("1. 助记词用来备份恢复钱包，请妥善安全保管。")]),a("p",[e._v("2. 密码只用来加密您的本地钱包数据。")])])],1)],1)],1)},k=[],C={data:function(){var e=this,t=function(t,a,n){""===a?n(new Error("请输入密码")):(""!==e.ruleForm.password2&&e.$refs.ruleForm.validateField("password2"),n())},a=function(t,a,n){""===a?n(new Error("请再次输入密码")):a!==e.ruleForm.password1?n(new Error("两次输入密码不一致!")):n()};return{ruleForm:{seed:"",mnemonic:"",password1:"",password2:""},rules:{password1:[{validator:t,trigger:"blur"}],password2:[{validator:a,trigger:"blur"}]}}},mounted:function(){this.$emit("checkWalletStats"),this.newSeed()},methods:{newSeed:function(){var e=this;this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"wallet_makeSeed",params:null})}).then(function(t){"undefined"!=typeof t.data.error?e.$alert("错误，请稍后重试","seed",{showClose:!1,confirmButtonText:"确定",callback:function(e){}}):(e.ruleForm.seed=t.data.result.seed,e.ruleForm.mnemonic=t.data.result.mnemonic)})},submitForm:function(e){var t=this,a=this;this.$refs.ruleForm.validate(function(e){if(!e)return!1;t.$emit("setLoading",!0,"创建钱包"),t.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"wallet_create",params:[t.ruleForm.seed,t.ruleForm.password1]})}).then(function(e){"undefined"==typeof e.data.error?t.$message({message:"\b创建成功!",type:"success",duration:500,onClose:function(){a.$emit("setLoading",!1,""),a.$emit("checkWalletStats")}}):t.$message({message:"错误，请稍后重试: "+e.data.error.message,type:"warning",duration:500,onClose:function(){a.$emit("setLoading",!1),a.$router.go(0)}})})})}}},F=C,E=Object(m["a"])(F,y,k,!1,null,null,null),O=E.exports,T=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",[a("el-col",{attrs:{span:6}},[a("h2",[e._v("恢复钱包")])]),a("el-col",{attrs:{span:6}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{ref:"ruleForm",attrs:{model:e.ruleForm,rules:e.rules,"label-width":"100px"}},[a("el-form-item",{attrs:{label:"助记词",prop:"mnemonic"}},[a("el-input",{attrs:{type:"textarea",autosize:{minRows:2,maxRows:4}},model:{value:e.ruleForm.mnemonic,callback:function(t){e.$set(e.ruleForm,"mnemonic",t)},expression:"ruleForm.mnemonic"}})],1),a("el-form-item",{attrs:{label:"请输入密码",prop:"password1"}},[a("el-input",{attrs:{placeholder:"请输入密码","show-password":""},model:{value:e.ruleForm.password1,callback:function(t){e.$set(e.ruleForm,"password1",t)},expression:"ruleForm.password1"}})],1),a("el-form-item",{attrs:{label:"再次输入密码",prop:"password2"}},[a("el-input",{attrs:{placeholder:"再次输入密码","show-password":""},model:{value:e.ruleForm.password2,callback:function(t){e.$set(e.ruleForm,"password2",t)},expression:"ruleForm.password2"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"},on:{click:function(t){return e.submitForm("ruleForm")}}},[e._v("恢复")])],1),a("div",[a("p",[e._v("注意：")]),a("p",[e._v("1. 助记词用来备份恢复钱包，请妥善安全保管。")]),a("p",[e._v("2. 密码只用来加密您的本地钱包数据。")])])],1)],1)],1)},A=[],S={data:function(){var e=this,t=function(t,a,n){""===a?n(new Error("请输入密码")):(""!==e.ruleForm.password2&&e.$refs.ruleForm.validateField("password2"),n())},a=function(t,a,n){""===a?n(new Error("请再次输入密码")):a!==e.ruleForm.password1?n(new Error("两次输入密码不一致!")):n()};return{ruleForm:{mnemonic:"",password1:"",password2:""},rules:{password1:[{validator:t,trigger:"blur"}],password2:[{validator:a,trigger:"blur"}]}}},mounted:function(){var e=this;this.$emit("checkWalletStats",function(t){e.$router.push("/")})},methods:{submitForm:function(e){var t=this,a=this;this.$refs.ruleForm.validate(function(e){if(!e)return!1;t.$emit("setLoading",!0,"恢复钱包"),t.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"wallet_recove",params:[t.ruleForm.mnemonic,t.ruleForm.password1]})}).then(function(e){"undefined"==typeof e.data.error?t.$message({message:"\b恢复成功成功!",type:"success",duration:500,onClose:function(){a.$emit("setLoading",!1,""),a.$emit("checkWalletStats",function(e){a.$router.push("/")})}}):t.$message({message:"错误，请稍后重试: "+e.data.error.message,type:"warning",duration:500,onClose:function(){a.$emit("setLoading",!1),a.$router.go(0)}})})})}}},j=S,L=Object(m["a"])(j,T,A,!1,null,null,null),R=L.exports,P=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex"}},[a("el-col",{attrs:{span:6}},[a("h2",[e._v("账号管理")])]),a("el-col",{attrs:{span:12}}),a("el-col",{attrs:{span:6}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newAccount}},[e._v("新建账号")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{key:Math.random(),attrs:{data:e.accountsTable}},[a("el-table-column",{attrs:{prop:"account",label:"名称",width:"120"}}),a("el-table-column",{attrs:{prop:"balance",label:"余额"}})],1)],1)],1)},N=[],W={data:function(){return{accountsTable:[{account:"defalut",balance:1}]}},methods:{listAccount2table:function(e){var t=[];for(var a in e)t.push({account:a,balance:e[a]});return t},newAccount:function(){this.$router.push({path:"/account/new"})}},mounted:function(){var e=this;this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_list",params:null})}).then(function(t){"undefined"==typeof t.data.error?(e.accountsTable=e.listAccount2table(t.data.result),e.$store.state.Accounts=e.accountsTable):e.$emit("alertResError",t.data.error,function(){e.$router.push("/")})})}},z=W,D=Object(m["a"])(z,P,N,!1,null,null,null),J=D.exports,M=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",[a("el-col",{attrs:{span:6}},[a("h2",[e._v("新建账号")])]),a("el-col",{attrs:{span:6}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{ref:"ruleForm",attrs:{model:e.ruleForm,rules:e.rules,"label-width":"100px"}},[a("el-form-item",{attrs:{label:"账号名称",prop:"name"}},[a("el-input",{attrs:{placeholder:"账号"},model:{value:e.ruleForm.name,callback:function(t){e.$set(e.ruleForm,"name",t)},expression:"ruleForm.name"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"},on:{click:function(t){return e.submitForm("ruleForm")}}},[e._v("创建")]),a("el-button",{on:{click:e.accountList}},[e._v("取消")])],1),a("div")],1)],1)],1)},B=[],U={data:function(){var e=function(e,t,a){"*"==t&&a(new Error("账号名不能为*")),a()};return{ruleForm:{account:""},rules:{account:[{validator:e,trigger:"blur"}]}}},mounted:function(){var e=this;e.$emit("checkWalletStats",function(t){"lock"==t&&e.$emit("dlgUnlockWallet",function(t){t||e.$router.push("/account/create")})})},methods:{submitForm:function(e){var t=this,a=this;this.$refs.ruleForm.validate(function(e){if(!e)return!1;t.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_create",params:[t.ruleForm.account]})}).then(function(e){"undefined"==typeof e.data.error?t.$alert("\b创建账号成功",{showClose:!1,closeOnClickModal:!1,closeOnPressEscape:!1,confirmButtonText:"确定",callback:function(e,a){t.$router.push({path:"/account"})}}):a.$emit("alertResError",e.data.error,function(){a.$router.push("/account")})})})},accountList:function(){this.$router.push({path:"/account"})},newAccount:function(){this.$router.push({path:"/account/new"})}}},Q=U,q=Object(m["a"])(Q,M,B,!1,null,null,null),I=q.exports,V=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex",justify:"space-between"}},[a("el-col",{attrs:{span:4}},[a("h2",[e._v("地址管理")])]),a("el-col",{attrs:{span:8}},[a("el-row",[a("el-col",{attrs:{span:8}},[a("span",{staticStyle:{color:"#303133","font-weight":"600"}},[e._v("当前账户：")])]),a("el-col",{attrs:{span:14}},[a("el-select",{attrs:{placeholder:"账号",size:"small"},on:{change:e.getAddressList},model:{value:e.current,callback:function(t){e.current=t},expression:"current"}},e._l(e.accounts,function(e){return a("el-option",{key:e.index,attrs:{label:e.account,value:e.account}})}),1)],1)],1)],1),a("el-col",{attrs:{span:4}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newAddress}},[e._v("新建地址")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.addresses}},[a("el-table-column",{attrs:{type:"index",width:"50"}}),a("el-table-column",{attrs:{prop:"addr",label:"地址"}})],1)],1)],1)},Y=[],H={data:function(){return{accounts:[],current:"",addresses:[]}},mounted:function(){0!=this.$store.state.Accounts.length?(this.accounts=this.$store.state.Accounts,this.current=this.$store.state.Accounts[0].account,this.getAddressList()):this.$router.push("/account")},methods:{newAddress:function(){var e=this,t=this;this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_createAddress",params:[t.current]})}).then(function(a){"undefined"==typeof a.data.error?e.addresses.push({addr:a.data.result}):t.$emit("alertResError",a.data.error,function(){})})},getAddressList:function(){var e=this;e.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_listAddresses",params:[e.current]})}).then(function(t){if("undefined"==typeof t.data.error){for(var a=[],n=0;n<t.data.result.length;n++)a.push({addr:t.data.result[n]});e.addresses=a}else e.$emit("alertResError",t.data.error,function(){})})}}},K=H,X=Object(m["a"])(K,V,Y,!1,null,null,null),Z=X.exports,G=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex",justify:"space-between"}},[a("el-col",{attrs:{span:4}},[a("h2",[e._v("发送交易")])]),a("el-col",{attrs:{span:8}},[a("el-row",[a("el-col",{attrs:{span:8}},[a("span",{staticStyle:{color:"#303133","font-weight":"600"}},[e._v("当前账户：")])]),a("el-col",{attrs:{span:14}},[a("el-select",{attrs:{placeholder:"账号",size:"small"},model:{value:e.current,callback:function(t){e.current=t},expression:"current"}},e._l(e.accounts,function(t){return a("el-option",{key:t.index,attrs:{label:t.account,value:t.index},on:{change:e.changeAccount}})}),1)],1)],1)],1),a("el-col",{attrs:{span:4}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{attrs:{model:e.form,"label-width":"120px"}},[a("el-form-item",{attrs:{label:"可用余额"}},[a("el-input",{attrs:{disabled:!0},model:{value:e.balance,callback:function(t){e.balance=t},expression:"balance"}})],1),a("el-form-item",{attrs:{label:"转给(to)"}},[a("el-input",{attrs:{placeholder:"address"},model:{value:e.form.to,callback:function(t){e.$set(e.form,"to",t)},expression:"form.to"}})],1),a("el-form-item",{attrs:{label:"转账金额(amount)"}},[a("el-input",{attrs:{placeholder:"amount"},model:{value:e.form.value,callback:function(t){e.$set(e.form,"value",t)},expression:"form.value"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"},on:{click:e.sentTx}},[e._v("创建")]),a("el-button",{on:{click:e.toAccount}},[e._v("取消")])],1)],1)],1)],1)},ee=[],te={data:function(){return{accounts:[],current:"",balance:"",form:{to:"",value:""}}},mounted:function(){0!=this.$store.state.Accounts.length?(this.accounts=this.$store.state.Accounts,this.current=this.$store.state.Accounts[0].account,this.balance=this.$store.state.Accounts[0].balance):this.$router.push("/account")},methods:{toAccount:function(){this.$router.push("/account")},changeAccount:function(e){this.balance=this.$store.state.Accounts[e].balance},sentTx:function(){var e=this,t=this;this.$axios({method:"post",data:JSON.stringify({id:(new Date).getTime(),method:"account_sendToAddress",params:[t.form.to,parseFloat(t.form.value),"",""]})}).then(function(a){e.$createElement;"undefined"==typeof a.data.error?e.$alert("\b创建账号成功",{showClose:!1,closeOnClickModal:!1,closeOnPressEscape:!1,confirmButtonText:"确定",callback:function(t,a){e.$router.push({path:"/account"})}}):t.$emit("alertResError",a.data.error,function(){})})}}},ae=te,ne=Object(m["a"])(ae,G,ee,!1,null,null,null),re=ne.exports,oe=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex",justify:"space-between"}},[a("el-col",{attrs:{span:4}},[a("h2",[e._v("交易记录")])]),a("el-col",{attrs:{span:8}},[a("el-row",[a("el-col",{attrs:{span:8}},[a("span",{staticStyle:{color:"#303133","font-weight":"600"}},[e._v("当前账户：")])]),a("el-col",{attrs:{span:14}},[a("el-select",{attrs:{placeholder:"账号",size:"small"},model:{value:e.current,callback:function(t){e.current=t},expression:"current"}},e._l(e.accounts,function(e){return a("el-option",{key:e.index,attrs:{label:e.account,value:e.index}})}),1)],1)],1)],1),a("el-col",{attrs:{span:4}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.txList}},[a("el-table-column",{attrs:{prop:"date",label:"时间",width:"160"}}),a("el-table-column",{attrs:{prop:"type",label:"-",width:"40"}}),a("el-table-column",{attrs:{prop:"to",label:"to",width:"240"}}),a("el-table-column",{attrs:{prop:"amount",label:"金额"}})],1)],1)],1)},le=[],se={data:function(){return{txList:[],accounts:[],current:""}},mounted:function(){0!=this.$store.state.Accounts.length?(this.accounts=this.$store.state.Accounts,this.current=this.$store.state.Accounts[0].account,this.getTxList()):this.$router.push("/account")},methods:{getTxList:function(){return[]}}},ce=se,ie=Object(m["a"])(ce,oe,le,!1,null,null,null),ue=ie.exports,de=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex"}},[a("el-col",{attrs:{span:6}},[a("h2",[e._v("备份恢复")])]),a("el-col",{attrs:{span:12}}),a("el-col",{attrs:{span:6}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newAccount}},[e._v("导入账号")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.tableData}},[a("el-table-column",{attrs:{prop:"alias",label:"别名",width:"120"}}),a("el-table-column",{attrs:{prop:"key",label:"key(encrypt)",width:"380"}}),a("el-table-column",[a("el-link",{attrs:{type:"primary",icon:"el-icon-download"}},[e._v("导出备份")]),e._v("  \n        "),a("el-link",{attrs:{type:"danger",icon:"el-icon-delete"}},[e._v("删除")])],1)],1)],1)],1)},me=[],pe={data:function(){var e=[{alias:"default",key:" tNNLc7iAQ5zzPeaHnz/UkqJrRezBSlpIzTMacSXRNhk=",balance:1231.12},{alias:"daodao",key:" 6evV6E7IOubFqBPF47N4TKoJcL4hUJYZlakYVgEi3Bo=",balance:123123123.12}];return{tableData:e}},methods:{newAccount:function(){this.$router.push({path:"/account/new"})}}},fe=pe,he=Object(m["a"])(fe,de,me,!1,null,null,null),we=he.exports,be=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",{attrs:{type:"flex"}},[a("el-col",{attrs:{span:6}},[a("h2",[e._v("节点管理")])]),a("el-col",{attrs:{span:12}}),a("el-col",{attrs:{span:6}},[a("el-button",{attrs:{type:"primary",icon:"el-icon-plus",size:"small"},on:{click:e.newNode}},[e._v("新建节点")])],1)],1)],1),a("el-main",{staticClass:"cmain"},[a("el-table",{attrs:{data:e.tableData,"highlight-current-row":""},on:{"current-change":e.handleCurrentChange}},[a("el-table-column",{attrs:{type:"index",width:"50"}}),a("el-table-column",{attrs:{prop:"name",label:"名称",width:"120"}}),a("el-table-column",{attrs:{prop:"addr",label:"地址"}}),a("el-table-column",{attrs:{prop:"user",label:"user",width:"120"}}),a("el-table-column",{attrs:{prop:"pwd",label:"pwd",width:"120"}}),a("el-table-column",[a("el-link",{attrs:{type:"danger",icon:"el-icon-delete",width:"80"}},[e._v("删除")])],1)],1)],1)],1)},ve=[],ge={data:function(){var e=[{name:"本地",addr:"127.0.0.1:1236",user:"admin",pwd:"123456"},{name:"daodao",addr:"12.4.23.4:1236",user:"admin",pwd:"123456"}];return{tableData:e}},methods:{newNode:function(){this.$router.push({path:"/node/new"})},handleCurrentChange:function(e){this.currentRow=e}}},$e=ge,_e=Object(m["a"])($e,be,ve,!1,null,null,null),xe=_e.exports,ye=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("el-container",[a("el-header",{staticClass:"cheader"},[a("el-row",[a("el-col",{attrs:{span:6}},[a("h2",[e._v("新建节点")])]),a("el-col",{attrs:{span:6}})],1)],1),a("el-main",{staticClass:"cmain"},[a("el-form",{attrs:{model:e.form,"label-width":"100px"}},[a("el-form-item",{attrs:{label:"名字"}},[a("el-input",{attrs:{placeholder:"nodename"},model:{value:e.form.name,callback:function(t){e.$set(e.form,"name",t)},expression:"form.name"}})],1),a("el-form-item",{attrs:{label:"RPC地址"}},[a("el-input",{attrs:{placeholder:"http://127.0.0.1:18130"},model:{value:e.form.addr,callback:function(t){e.$set(e.form,"addr",t)},expression:"form.addr"}})],1),a("el-form-item",{attrs:{label:"RPC用户名"}},[a("el-input",{attrs:{placeholder:"RPC user"},model:{value:e.form.user,callback:function(t){e.$set(e.form,"user",t)},expression:"form.user"}})],1),a("el-form-item",{attrs:{label:"RPC密码"}},[a("el-input",{attrs:{placeholder:"RPC password"},model:{value:e.form.password,callback:function(t){e.$set(e.form,"password",t)},expression:"form.password"}})],1),a("el-form-item",[a("el-button",{attrs:{type:"primary"}},[e._v("创建")]),a("el-button",{on:{click:e.nodeList}},[e._v("取消")])],1)],1)],1)],1)},ke=[],Ce={data:function(){return{form:{name:"",addr:"",user:"",password:""}}},methods:{nodeList:function(){this.$router.push({path:"/node"})}}},Fe=Ce,Ee=Object(m["a"])(Fe,ye,ke,!1,null,null,null),Oe=Ee.exports;n["default"].use(w["a"]);var Te=new w["a"]({routes:[{path:"/",name:"index",component:x},{path:"/wallet/create",name:"walletcreate",component:O},{path:"/wallet/recove",name:"walletrecove",component:R},{path:"/account",name:"account",component:J},{path:"/account/new",name:"accountnew",component:I},{path:"/address",name:"address",component:Z},{path:"/tx/send",name:"txsend",component:re},{path:"/tx/list",name:"txlist",component:ue},{path:"/backup",name:"backup",component:we},{path:"/node",name:"node",component:xe},{path:"/node/new",name:"nodenew",component:Oe}]}),Ae=a("2f62");n["default"].use(Ae["a"]);var Se={Accounts:[]},je=new Ae["a"].Store({state:Se}),Le=je;n["default"].use(o.a),n["default"].prototype.$axios=s.a.create({baseURL:window.QitmeerConfig.RPCAddr,timeout:5e3,headers:{"content-type":"application/json"},auth:{username:window.QitmeerConfig.RPCUser,password:window.QitmeerConfig.RPCPass},withCredentials:!0,crossDomain:!0}),new n["default"]({el:"#app",router:Te,store:Le,render:function(e){return e(f)}})},"845fb":function(e,t,a){},be35:function(e,t,a){},cf05:function(e,t,a){e.exports=a.p+"assets/img/logo.c9da5caa.png"}});
//# sourceMappingURL=app.69a6959b.js.map