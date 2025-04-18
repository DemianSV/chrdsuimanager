<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.rawdata')" to="/rawdata" disabled />
        </va-breadcrumbs>
    </div>

    <div class="layout va-gutter-4 md12" style="height: 100%; width: 100%">
        <div class="row">
            <va-select class="flex md6" v-model="spaceValue" :options="spaceOptions" v-on:update:model-value="spaceSelectUpdateValue" :label="$t('rawdata.filter01')" :placeholder="$t('rawdata.message01')" :no-options-text="$t('message.listempty')"></va-select>
            <va-select class="flex md6" v-model="metricValue" :options="metricOptions" v-on:update:model-value="metricSelectUpdateValue" :label="$t('rawdata.filter02')" :placeholder="$t('rawdata.message02')" :no-options-text="$t('message.listempty')"></va-select>
        </div>
        <div class="row align-end">
            <va-input class="flex md6" :placeholder="$t('rawdata.message03')" :label="$t('rawdata.filter03')" v-model="input" v-on:update:model-value="filterUpdate"></va-input>
            <div class="flex md6">
                <va-pagination class="justify-center" v-model="pageCurent" size="small" :visible-pages="10" :pages="pageCount" v-on:update:model-value="pageUpdateValue" boundary-numbers></va-pagination>
            </div>
        </div>
        <br>
        <va-inner-loading :loading="tableLoading">
            <ag-grid-vue :key="componentKey" :rowData="tableData" :columnDefs="colDefs" domLayout="autoHeight" @grid-ready="onGridReady" style="height: 100%; width: 100%"></ag-grid-vue>
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
                spaceValue: {},
                spaceOptions: [],
                metricOptions: [],
                metricValue: {},
                pageCurent: 1,
                pageCount: 1,
                gridApi: null,
                tableData: [],
                tableLoading: false,
                dateMin: this.$t('message.nodata'),
                dateMax: this.$t('message.nodata'),
                input: "",
                wrapperSize: "800px",
                componentKey: 0,
                colDefs: [
                    { field: "spacedesc", headerName: this.$t('rawdata.table01') },
                    { field: "object", headerName: this.$t('rawdata.table02') },
                    { field: "metric", headerName: this.$t('rawdata.table03') },
                    { field: "value", headerName: this.$t('rawdata.table04') },
                    { field: "eventtime", headerName: this.$t('rawdata.table05') },
                    { field: "status", headerName: this.$t('rawdata.table06') },
                ],
            }
        },
        methods: {
            rawDataSelect() {
                const vm = this;
                vm.tableLoading = true;
                if (this.spaceValue.value != "") {
                    if (this.metricValue.value != "") {
                        let dataPut = {
                            spaceid: this.spaceValue.value,
                            metric: this.metricValue.value,
                            pagecurent: this.pageCurent,
                        };
                        $.ajax({
                            url: "/api/v1/admin/rawdata/select?" + Math.random(),
                            type: "PUT",
                            dataType: "json",
                            data: JSON.stringify(dataPut),
                            statusCode: {
                                200: function (data) {
                                    let dataNormal = [];
                                    if (data.data != null) {
                                        data.data.sort(function(a, b) {
                                            return b.eventtime - a.eventtime;
                                        });
                                        data.data.forEach(function(item, i) {
                                            dataNormal[i] = item;
                                            dataNormal[i].createtime = moment(item.createtime).format("DD.MM.YYYY HH:mm:ss");
                                            dataNormal[i].eventtime = moment(item.eventtime).format("DD.MM.YYYY HH:mm:ss");
                                        });
                                    }
                                    vm.tableData = dataNormal;
                                    vm.pageCount = data.page.pagecount;
                                    vm.pageCurent = data.page.pagecurent;
                                    if (data.page.datemin > 0 && data.page.datemax > 0) {
                                        vm.dateMin = moment(data.page.datemin).format("DD.MM.YYYY HH:mm:ss");
                                        vm.dateMax = moment(data.page.datemax).format("DD.MM.YYYY HH:mm:ss");
                                    }
                                    vm.tableLoading = false;
                                    return true;
                                },
                                500: function () {
                                    vm.tableLoading = false;
                                    return true;
                                }
                            },
                            error: function () {
                                vm.tableLoading = false;
                                return true;
                            }
                        });
                    } else {
                        vm.tableLoading = false;
                    }
                } else {
                    vm.tableLoading = false;
                }
            },
            spaceSelect() {
                const vm = this;
                $.ajax({
                    url: "/api/v1/admin/space/select?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        if (data == null) {
                            data = [];
                        } else {
                            if (data.length > 0) {
                                data.forEach(function(item, i) {
                                    vm.spaceOptions[i] = { text: item.description, value: item.id };
                                });
                            }
                        }
                        return true;
                    }
                });
            },
            spaceSelectUpdateValue() {
                const vm = this;
                vm.metricOptions = [];
                vm.metricValue = {};
                let dataPut = {
                    spaceid: this.spaceValue.value,
                };
                $.ajax({
                    url: "/api/v1/admin/rawdata/metric/select?" + Math.random(),
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    success: function (data) {
                        if (data == null) {
                            data = [];
                        } else {
                            if (data.length > 0) {
                                data.forEach(function(item, i) {
                                    vm.metricOptions[i + 1] = {text: item.metric, value: item.metric};
                                });
                            }
                        }
                        return true;
                    }
                });
                if (vm.spaceValue.value != "") {
                    vm.tableData = [];
                    vm.pageCount = 1;
                    vm.pageCurent = 1;
                }
            },
            pageUpdateValue() {
                this.rawDataSelect();
            },
            metricSelectUpdateValue() {
                this.input = "";
                if (this.metricValue.value != "") {
                    this.rawDataSelect();
                }
            },
            filterUpdate() {
                this.gridApi.setGridOption("quickFilterText", this.input);
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
            this.spaceSelect();
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
                    { field: "spacedesc", headerName: this.$t('rawdata.table01') },
                    { field: "object", headerName: this.$t('rawdata.table02') },
                    { field: "metric", headerName: this.$t('rawdata.table03') },
                    { field: "value", headerName: this.$t('rawdata.table04') },
                    { field: "eventtime", headerName: this.$t('rawdata.table05') },
                    { field: "status", headerName: this.$t('rawdata.table06') },
                ]
                this.componentKey += 1;
            },
            mode(newValue, oldValue) {
                this.componentKey += 1;
            },
            tableData(newValue) {
                if (newValue.length > 0 && this.gridApi) {
                    console.log("SIZE COLUMNS");
                    setTimeout(() => {
                        // this.gridApi.autoSizeAllColumns();
                        this.gridApi.sizeColumnsToFit();
                    }, 50);
                }
            },
        }
    }
</script>

<style>
    .va-button-group {
        display: block;
        padding: 0.75rem;
    }
</style>