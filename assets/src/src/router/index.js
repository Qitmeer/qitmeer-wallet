import Vue from 'vue';
import Router from 'vue-router';

import index from '@/components/index';

import walletcreate from '@/components/walletcreate';
import walletrecover from '@/components/walletrecover';
import account from '@/components/account';
import accountnew from '@/components/accountnew';
import address from '@/components/address';
import txsend from '@/components/txsend';
import txlist from '@/components/txlist';
import backup from "@/components/backup";
import backupImport from "@/components/importkey";

import node from "@/components/node";
import nodenew from "@/components/nodenew";

Vue.use(Router)

export default new Router({
    routes: [
        {
            path: '/',
            name: 'index',
            component: index
        },
        {
            path: '/wallet/create',
            name: 'walletcreate',
            component: walletcreate
        },
        {
            path: '/wallet/recover',
            name: 'walletrecover',
            component: walletrecover
        },
        {
            path: '/account',
            name: 'account',
            component: account
        },
        {
            path: '/account/new',
            name: 'accountnew',
            component: accountnew
        },
        {
            path: '/address',
            name: 'address',
            component: address
        },
        {
            path: '/tx/send',
            name: 'txsend',
            component: txsend
        },
        {
            path: '/tx/list',
            name: 'txlist',
            component: txlist
        },
        {
            path: '/backup',
            name: 'backup',
            component: backup
        },
        {
            path: '/backup/import',
            name: 'import',
            component: backupImport
        },
        {
            path: '/node',
            name: 'node',
            component: node
        },
        {
            path: '/node/new',
            name: 'nodenew',
            component: nodenew
        },
        {
            path: '/node/edit/:name',
            name: 'nodeedit',
            component: nodenew
        }
    ]
})