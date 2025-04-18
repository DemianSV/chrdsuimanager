<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.problems')" to="/problem" disabled />
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
    import moment from 'moment'
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
                tableData: [],
                moduleid: "",
                moduledesc: "",
                spaceid: "",
                spacedesc: "",
                metric: "",
                eventtime: "",
                eventtimestart: "",
                status: "",
                value: "",
                tableLoading: false,
                gridApi: null,
                componentKey: 0,
                colDefs: [
                    { field: "spacedesc", headerName: this.$t('problem.table01') },
                    { field: "moduledesc", headerName: this.$t('problem.table02') },
                    { field: "metric", headerName: this.$t('problem.table03') },
                    { field: "eventtime", headerName: this.$t('problem.table04') },
                    { field: "eventtimestart", headerName: this.$t('problem.table05') },
                    { field: "value", headerName: this.$t('problem.table06') },
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
                if (params.value === "NORMAL") {
                    return "problem-green";
                } else if (params.value === "WARNING") {
                    return "problem-yellow";
                } else if (params.value === "CRITICAL") {
                    return "problem-red";
                }
            },
            problemSelect() {
                const vm = this;
                vm.tableLoading = true;
                $.ajax({
                    url: "/api/v1/admin/problem/select?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        let dataNormal = [];
                        if (data != null) {
                            data.forEach(function(item, i) {
                                dataNormal[i] = item;
                                dataNormal[i].eventtime = moment(item.eventtime).format("DD.MM.YYYY HH:mm:ss");
                                dataNormal[i].eventtimestart = moment(item.eventtimestart).format("DD.MM.YYYY HH:mm:ss");
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
                        if (data.role == "superadmin" || data.role == "admin") {
                            vm.userBlock = false;
                        } else {
                            vm.userBlock = false;
                        }
                        return true;
                    }
                });
                this.problemSelect();
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
                    { field: "spacedesc", headerName: this.$t('problem.table01') },
                    { field: "moduledesc", headerName: this.$t('problem.table02') },
                    { field: "metric", headerName: this.$t('problem.table03') },
                    { field: "eventtime", headerName: this.$t('problem.table04') },
                    { field: "eventtimestart", headerName: this.$t('problem.table05') },
                    { field: "value", headerName: this.$t('problem.table06') },
                    {
                        field: "status", 
                        headerName: this.$t('message.status'),
                        cellClass: this.cellClassStatus,
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
        },
    }
</script>

<style>
.problem-red {
    background-color: rgb(224, 80, 61);
    color: black;
}
.problem-yellow {
    background-color: rgb(249, 218, 106);
    color: black;
}
.problem-green {
    background-color: rgb(102, 190, 51);
    color: black;
}
</style>