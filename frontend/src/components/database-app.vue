<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.dbstatus')" to="/database" disabled />
        </va-breadcrumbs>
    </div>

<div class="layout va-gutter-4 md12">
    <va-inner-loading :loading="tableLoading">
        <ag-grid-vue :key="componentKey" :rowData="tableData" :columnDefs="colDefs" @grid-ready="onGridReady" domLayout="autoHeight" style="height: 100%; width: 100%"></ag-grid-vue>
    </va-inner-loading>
</div>
</template>

<script>
import $ from 'jquery'
import { mapState } from 'vuex';
import { AgGridVue } from "ag-grid-vue3"; // Vue Data Grid Component

export default {
    components: {
        AgGridVue, // Add Vue Data Grid component
    },
    data() {
        return {
            userRole: null,
            userBlock: true,
            adminBlock: true,
            tableData: [],
            tableLoading: false,
            gridApi: null,
            componentKey: 0,
            colDefs: [
                    { field: "peer", headerName: "Peer" },
                    { field: "datacenter", headerName: "DC" },
                    { field: "hostid", headerName: "Host ID" },
                    { field: "owns", headerName: "Owns" },
                    { field: "tokens", headerName: "Tokens" },
                    { field: "load", headerName: "Load" },
                    {
                        field: "up",
                        headerName: "Up",
                        cellClass: this.cellClassUP,
                    },
                    {
                        field: "status", 
                        headerName: this.$t('message.status'),
                        cellClass: this.cellClassStatus,
                    },
                ],
        }
    },
    methods: {
        cellClassStatus(params) {
            return params.value === "NORMAL" ? "db-green" : "db-red";
        },
        cellClassUP(params) {
            return params.value === "UP" ? "db-green" : "db-red";
        },
        getUpText(value) {
            if (value === true) {
                return "UP";
            } else {
                return "DOWN";
            }
        },
        databaseStatus() {
            const vm = this;
            vm.tableLoading = true;
            $.ajax({
                url: "/api/v1/admin/database/status?" + Math.random(),
                type: "GET",
                dataType: "json",
                success: function (data) {
                    let dataNormal = [];
                    if (data != null) {
                        data.forEach(function(item, i) {
                            dataNormal[i] = item;
                            dataNormal[i].datacenter = (item.datacenter).toUpperCase();
                            dataNormal[i].status = (item.status).toUpperCase();
                            dataNormal[i].up = vm.getUpText(item.up)
                        });
                    }
                    
                    vm.tableData = dataNormal;
                    vm.tableLoading = false;
                    return true;
                }
            });
        },
        async userInfo() {
            const vm = this;
            await $.ajax({
                url: "/api/v1/userinfo?" + Math.random(),
                type: "GET",
                dataType: "json",
                success: function (data) {
                    vm.userRole = data.role;
                    if (data.role == "superadmin") {
                        vm.userBlock = false
                        vm.adminBlock = false
                    } else if (data.role == "admin") {
                        vm.userBlock = false
                        vm.adminBlock = true
                        vm.databaseData = "Нет доступа"
                    }
                    else {
                        vm.userBlock = true
                        vm.adminBlock = true
                        vm.databaseData = "Нет доступа"
                        vm.userData = "Нет доступа"
                    }
                    return true;
                }
            });
            console.log("Роль: " + this.userRole);
            console.log("Ограничение: " + this.userBlock + ", " + this.adminBlock);
            if (this.userRole == "superadmin") {
                this.databaseStatus();
            }
        },
        onGridReady(params) {
            this.gridApi = params.api;
            this.gridApi.sizeColumnsToFit();
        },
        handleResize() {
            if (this.gridApi) {
                this.gridApi.sizeColumnsToFit();
            }
        },
    },
    mounted() {
        window.addEventListener('resize', this.handleResize);
        this.userInfo();
    },
    beforeUnmount() {
        window.removeEventListener('resize', this.handleResize);
    },
    computed: {
            ...mapState(['locale', 'mode'])
        },
    watch: {
        locale(newValue, oldValue) {
            this.colDefs = [
                { field: "peer", headerName: "Peer" },
                { field: "datacenter", headerName: "DC" },
                { field: "hostid", headerName: "Host ID" },
                { field: "owns", headerName: "Owns" },
                { field: "tokens", headerName: "Tokens" },
                { field: "load", headerName: "Load" },
                {
                    field: "up",
                    headerName: "Up",
                    cellClass: this.cellClassUP,
                },
                {
                    field: "status", 
                    headerName: this.$t('message.status'),
                    cellClass: this.cellClass,
                },
            ];
            this.componentKey += 1;
        },
        mode(newValue, oldValue) {
            this.componentKey += 1;
        },
        tableData(newValue) {
            if (newValue.length > 0 && this.gridApi) {
                setTimeout(() => {
                    this.gridApi.sizeColumnsToFit();
                }, 50);
            }
        },
    }
}
</script>

<style>
.db-red {
    background-color: rgb(224, 80, 61);
    color: black;
}
.db-green {
    background-color: rgb(102, 190, 51);
    color: black;
}
</style>