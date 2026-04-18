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
                    { field: "host", headerName: "Host" },
                    { field: "dc", headerName: "DC" },
                    { field: "rack", headerName: "Rack" },
                    { field: "version", headerName: "Version" },
                    { field: "stat", headerName: "Stat" },
                ],
        }
    },
    methods: {
        databaseStatus() {
            let vm = this;
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
                            dataNormal[i].dc = (item.dc).toUpperCase();
                            dataNormal[i].rack = (item.rack).toUpperCase();
                            dataNormal[i].stat = (item.stat).toUpperCase();
                        });
                    }
                    
                    vm.tableData = dataNormal;
                    vm.tableLoading = false;
                    return true;
                }
            });
        },
        async userInfo() {
            let vm = this;
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
            if (this.userRole == "superadmin" || this.userRole == "admin" || this.userRole == "user") {
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
                    { field: "host", headerName: "Host" },
                    { field: "dc", headerName: "DC" },
                    { field: "rack", headerName: "Rack" },
                    { field: "version", headerName: "Version" },
                    { field: "stat", headerName: "Stat" },
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