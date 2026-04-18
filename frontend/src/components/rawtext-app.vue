<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.rawtext')" to="/rawtext" disabled />
        </va-breadcrumbs>
    </div>
    
    <div class="layout va-gutter-4 md12">
        <div class="row">
            <va-select class="flex md6" v-model="spaceValue" :options="spaceOptions" v-on:update:model-value="spaceSelectUpdateValue" :label="$t('rawtext.filter01')" :placeholder="$t('rawtext.message01')" :no-options-text="$t('message.listempty')" searchable></va-select>
            <va-select class="flex md6" v-model="metricValue" :options="metricOptions" v-on:update:model-value="metricSelectUpdateValue" :label="$t('rawtext.filter02')" :placeholder="$t('rawtext.message02')" :no-options-text="$t('message.listempty')" searchable></va-select>
        </div>
        <div class="row align-end">
            <va-input class="flex md6" :placeholder="$t('rawtext.message03')" :label="$t('rawtext.filter03')" v-model="input" v-on:update:model-value="filterUpdate"></va-input>
            <div class="flex md6" align="center">
                <va-button icon="chevron_left" v-on:click="selectLeft()" v-bind:disabled="isDisabledLeft" class="mr-2 ml-1" />
                <va-button disabled>{{ currentPage + 1 }}</va-button>
                <va-button icon="chevron_right" v-on:click="selectRight()" v-bind:disabled="isDisabledRight"  class="ml-2" />
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
                tableData: [],
                tableLoading: false,
                input: "",
                filter: "",
                wrapperSize: "800px",
                componentKey: 0,
                gridApi: null,
                colDefs: [
                    { field: "spacedesc", headerName: this.$t('rawtext.table01') },
                    { field: "object", headerName: this.$t('rawtext.table02') },
                    { field: "metric", headerName: this.$t('rawtext.table03') },
                    { field: "value", headerName: this.$t('rawtext.table04') },
                    { field: "eventtime", headerName: this.$t('rawtext.table05') },
                ],

                // Pagination
                selectParam: {},
                isDisabledLeft: true,
                isDisabledRight: true,
                currentPage: 0,
                pages: [""],
            }
        },
        methods: {
            selectRight() {
                let vm = this;
                vm.currentPage++;
                vm.rawTextSelect(vm.selectParam.next);
                vm.pages[vm.currentPage] = vm.selectParam.next;
            },
            selectLeft() {
                let vm = this;
                if (vm.currentPage > 0) {
                    vm.currentPage--;
                }
                vm.rawTextSelect(vm.pages[vm.currentPage]);
            },
            rawTextSelect(current) {
                let vm = this;
                let dataPut = {};

                vm.tableLoading = true;

                if (current != "") {
                    dataPut.current = current;
                } else {
                    dataPut.current = "";
                }
                dataPut.pagesize = 1000;

                if (this.spaceValue.value != "") {
                    if (this.metricValue.value != "") {
                        dataPut.spaceid = this.spaceValue.value;
                        dataPut.metric = this.metricValue.value;

                        $.ajax({
                            url: "/api/v1/admin/rawtext/select?" + Math.random(),
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
                                            if (item.labels != null) {
                                                dataNormal[i].labels = JSON.stringify(item.labels);
                                            } else {
                                                dataNormal[i].labels = "";
                                            }
                                        });
                                    }
                                    vm.tableData = dataNormal;
                                    vm.selectParam.current = data.current;
                                    vm.selectParam.next = data.next;

                                    if (data.next != "" && data.next != undefined ) {
                                        vm.isDisabledRight = false;
                                    } else {
                                        vm.isDisabledRight = true;
                                    }
                                    if (data.current != "" && data.current != undefined) {
                                        vm.isDisabledLeft = false;
                                    } else {
                                        vm.isDisabledLeft = true;
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
                    }
                }
            },
            spaceSelect() {
                let vm = this;
                let dataPut = {
                    pagesize: 0,
                };

                $.ajax({
                    url: "/api/v1/admin/space/select?" + Math.random(),
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    success: function (data) {
                        if (data.data == null) {
                            data.data = [];
                        } else {
                            if (data.data.length > 0) {
                                data.data.forEach(function(item, i) {
                                    vm.spaceOptions[i] = { text: item.description, value: item.id };
                                });
                            }
                        }
                        return true;
                    }
                });
            },
            spaceSelectUpdateValue() {
                let vm = this;
                vm.metricOptions = [];
                vm.metricValue = {};

                let dataPut = {
                    spaceid: this.spaceValue.value,
                };
                $.ajax({
                    url: "/api/v1/admin/rawtext/metric/select?" + Math.random(),
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
                this.rawTextSelect();
            },
            metricSelectUpdateValue() {
                this.input = "";
                this.filter = this.metricValue.value;
                if (this.metricValue.value != "") {
                    this.selectParam = {};
                    this.isDisabledLeft = true;
                    this.isDisabledRight = true;
                    this.currentPage = 0;
                    this.pages = [""];
                    this.rawTextSelect();
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
                    { field: "spacedesc", headerName: this.$t('rawtext.table01') },
                    { field: "object", headerName: this.$t('rawtext.table02') },
                    { field: "metric", headerName: this.$t('rawtext.table03') },
                    { field: "value", headerName: this.$t('rawtext.table04') },
                    { field: "eventtime", headerName: this.$t('rawtext.table05') },
                ]
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

<style></style>