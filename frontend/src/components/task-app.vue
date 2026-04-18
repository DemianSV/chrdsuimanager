<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.tasks')" to="/task" disabled />
        </va-breadcrumbs>
    </div>

    <div class="row layout va-gutter-3">
        <div class="flex layout va-gutter-3 md6" align="left">
            <va-button icon="chevron_left" v-on:click="selectLeft()" v-bind:disabled="isDisabledLeft" class="mr-2 ml-1" />
            <va-button disabled>{{ currentPage + 1 }}</va-button>
            <va-button icon="chevron_right" v-on:click="selectRight()" v-bind:disabled="isDisabledRight"  class="ml-2" />
        </div>
        <div class="flex layout va-gutter-3 md6" align="right">
            <va-button icon="refresh" v-on:click="putTaskSelect(pages[currentPage])" class="ml-2 mr-1" />
            <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickNew" class="ml-2 mr-1" />
        </div>
    </div>
    
    <div class="row layout va-gutter-4 md12">
        <va-inner-loading :loading="tableLoading">
            <ag-grid-vue :key="componentKey" :rowData="tableData" :columnDefs="colDefs" domLayout="autoHeight" :components="components" @grid-ready="onGridReady" style="height: 100%; width: 100%" :enableCellTextSelection="true" :ensureDomOrder="true"></ag-grid-vue>
        </va-inner-loading>
    </div>
    
    <va-modal v-model="showModal02" :title="$t('task.new')" v-on:ok="putTaskCreate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-select class="mb-4" v-model="moduleValue" :options="moduleOptions" v-on:update:model-value="isErrorCh" :label="$t('message.module')" :placeholder="$t('task.message02')" :no-options-text="$t('message.listempty')" v-bind:error="isError01"></va-select></div>
        <div><va-select class="mb-4" v-model="spaceValue" :options="spaceOptions" v-on:update:model-value="isErrorCh" :label="$t('message.space')" :placeholder="$t('task.message01')" :no-options-text="$t('message.listempty')" v-bind:error="isError02"></va-select></div>
        <div><va-select class="mb-4" v-model="objectValue" :placeholder="$t('task.message03')" :label="$t('task.message04')" v-on:update:model-value="isErrorCh" v-model:options="objectOptions" max-height="150px" v-bind:error="isError03" @update:modelValue="objectUpdate()"></va-select></div>
        <div><va-select class="mb-4" v-model="dmValue" :options="dmOptions" v-on:update:model-value="isErrorCh" :label="$t('message.dm')" :placeholder="$t('task.message24')" :no-options-text="$t('message.listempty')" v-bind:disabled="isDisabledDM" clearable></va-select></div>
        <div><va-input class="mb-4" v-model="metric" :placeholder="$t('task.message05')" v-on:update:model-value="isErrorCh" :label="$t('task.message06')" v-bind:error="isError04"></va-input></div>
        <div><va-select class="mb-4" v-model="datatypeValue" :options="datatypeOptions" :label="$t('task.message07')" :placeholder="$t('task.message08')" :no-options-text="$t('message.listempty')"></va-select></div>
        <div><va-input class="mb-4" v-model="interval" :placeholder="$t('task.message09')" :label="$t('task.message10')" v-bind:disabled="isDisabled"></va-input></div>
        <div><va-input class="mb-4" background="danger" v-model="critical" :placeholder="$t('task.message11')" :label="$t('task.message12')"></va-input></div>
        <div><va-input class="mb-4" background="warning" v-model="warning" :placeholder="$t('task.message13')" :label="$t('task.message14')"></va-input></div>
        <div><va-select class="mb-4" v-model="emailListValue" :options="emailListOptions" v-on:update:model-value="isErrorCh" :label="$t('task.message22')" :placeholder="$t('task.message23')" :no-options-text="$t('message.listempty')" clearable></va-select></div>
        <div><va-select v-model="statusValue" class="mb-4" :placeholder="$t('task.message15')" :label="$t('message.status')" v-model:options="statusOptions"></va-select></div>
    </va-modal>
    <va-modal v-model="showModal01" :title="$t('task.message21')" v-on:ok="putTaskUpdate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-select class="mb-4" v-model="moduleValue" :options="moduleOptions" v-on:update:model-value="isErrorCh" :label="$t('message.module')" :placeholder="$t('task.message02')" :no-options-text="$t('message.listempty')" v-bind:error="isError01" disabled></va-select></div>
        <div><va-select class="mb-4" v-model="spaceValue" :options="spaceOptions" v-on:update:model-value="isErrorCh" :label="$t('message.space')" :placeholder="$t('task.message01')" :no-options-text="$t('message.listempty')" v-bind:error="isError02" disabled></va-select></div>
        <div><va-select class="mb-4" v-model="objectValue" :placeholder="$t('task.message03')" :label="$t('task.message04')" v-on:update:model-value="isErrorCh" v-model:options="objectOptions" max-height="150px" v-bind:error="isError03" @update:modelValue="objectUpdate()"></va-select></div>
        <div><va-select class="mb-4" v-model="dmValue" :options="dmOptions" v-on:update:model-value="isErrorCh" :label="$t('message.dm')" :placeholder="$t('task.message24')" :no-options-text="$t('message.listempty')" v-bind:disabled="isDisabledDM" clearable></va-select></div>
        <div><va-input class="mb-4" v-model="metric" :placeholder="$t('task.message05')" :label="$t('task.message06')" disabled></va-input></div>
        <div><va-select class="mb-4" v-model="datatypeValue" :options="datatypeOptions" :label="$t('task.message07')" :placeholder="$t('task.message08')" :no-options-text="$t('message.listempty')"></va-select></div>
        <div><va-input class="mb-4" v-model="interval" :placeholder="$t('task.message09')" :label="$t('task.message10')" v-bind:disabled="isDisabled"></va-input></div>
        <div><va-input class="mb-4" background="danger" v-model="critical" :placeholder="$t('task.message11')" :label="$t('task.message12')"></va-input></div>
        <div><va-input class="mb-4" background="warning" v-model="warning" :placeholder="$t('task.message13')" :label="$t('task.message14')"></va-input></div>
        <div><va-select class="mb-4" v-model="emailListValue" :options="emailListOptions" v-on:update:model-value="isErrorCh" :label="$t('task.message22')" :placeholder="$t('task.message23')" :no-options-text="$t('message.listempty')" clearable></va-select></div>
        <div><va-select v-model="statusValue" class="mb-4" :placeholder="$t('task.message15')" :label="$t('message.status')" v-model:options="statusOptions"></va-select></div>
    </va-modal>
    <va-modal v-model="showModal03" :title="$t('task.message16')" v-on:ok="putTaskRemove(rowIndexRemove)" :okText="$t('message.delete')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div>Подтвердите уаление задачи {{ taskIDRemove }}</div>
    </va-modal>
</template>

<script>
    import $ from 'jquery'
    import { mapState } from 'vuex';
    import { AgGridVue } from "ag-grid-vue3"; // Vue Data Grid Component
    import ActionRenderer from './ag-grid-app-action.vue'

    export default {
        components: {
            AgGridVue,
        },
        data() {
            const statusOptions = [
                    {text: this.$t('message.on'), value: 1},
                    {text: this.$t('message.off'), value: 0},
                ];
            const datatypeOptions = [
                    {text: this.$t('message.number'), value: "float"},
                    {text: this.$t('message.text'), value: "text"},
                ];
            const objectOptions = [
                    {text: "zbxagent", value: "zbxagent"},
                    {text: "prometheus", value: "prometheus"},
                    {text: "zbxagentpassive", value: "zbxagentpassive"},
                ];

            return {
                isError01: false,
                isError02: false,
                isError03: false,
                isError04: false,
                isDisabled: false,
                isDisabledDM: true,
                spaceOptions: [],
                moduleOptions: [],
                dmOptions: [],
                emailListOptions: [],
                emailListValue: {},
                spaceValue: {},
                moduleValue: {},
                dmValue: {},
                userRole: null,
                userBlock: true,
                rowIndexRemove: null,
                taskModuleIDRemove: "",
                taskSpaceIDRemove: "",
                taskObjectRemove: "",
                taskMetricRemove: "",
                noData: "Нет данных",
                showModal01: false,
                showModal02: false,
                showModal03: false,
                tableData: [],
                id: "",
                moduleid: "",
                spaceid: "",
                metric: "",
                critical: "",
                warning: "",
                interval: "",
                datatype: "",
                moduledesc: "",
                spacedesc: "",
                datatypeOptions: datatypeOptions,
                datatypeValue: datatypeOptions[0],
                statusOptions: statusOptions,
                statusValue: statusOptions[0],
                objectOptions: objectOptions,
                objectValue: objectOptions[0],
                tableLoading: false,
                gridApi: null,
                componentKey: 0,

                // Pagination
                selectParam: {},
                isDisabledLeft: true,
                isDisabledRight: true,
                currentPage: 0,
                pages: [""],

                // Column Definitions: Defines the columns to be displayed.
                colDefs: [
                    { field: "moduledesc", headerName: this.$t('message.module') },
                    { field: "spacedesc", headerName: this.$t('message.space') },
                    { field: "object", headerName: this.$t('task.object') },
                    { field: "metric", headerName: this.$t('task.metric') },
                    { field: "interval", headerName: this.$t('task.interval') },
                    { field: "status", headerName: this.$t('message.status') },
                    { 
                        field: "action",
                        headerName: this.$t('message.action'),
                        cellRenderer: 'actionRenderer',
                        cellRendererParams: {
                            rowClickEdit: this.rowClickEdit,
                            rowClickRemove: this.rowClickRemove,
                            userBlock: this.userBlock,
                        },
                    },
                ],
                components: {
                    actionRenderer: ActionRenderer,
                },
            }
        },
        methods: {
            objectUpdate() {
                if (this.objectValue.value == "prometheus") {
                    this.isDisabled = true;
                    this.isDisabledDM = false;
                } else {
                    this.isDisabled = false;
                    this.isDisabledDM = true;
                    this.dmValue = {};
                }
            },
            isErrorCh() {
                this.isError01 = false;
                this.isError02 = false;
                this.isError03 = false;
                this.isError04 = false;
            },
            rowClickEdit(value) {
                let vm = this;

                this.emailListValue = "";

                this.moduleOptions.forEach(function(item) {
                    if (item.value == vm.tableData[value].moduleid) {
                        vm.moduleValue = item;
                        return;
                    }
                })
                this.dmOptions.forEach(function(item) {
                    if (item.value == vm.tableData[value].dmid) {
                        vm.dmValue = item;
                        return;
                    }
                })
                this.spaceOptions.forEach(function(item) {
                    if (item.value == vm.tableData[value].spaceid) {
                        vm.spaceValue = item;
                        return;
                    }
                })
                this.statusOptions.forEach(function(item) {
                    if (item.value == vm.tableData[value].status) {
                        vm.statusValue = item;
                        return;
                    }
                })
                this.datatypeOptions.forEach(function(item) {
                    if (item.value == vm.tableData[value].datatype) {
                        vm.datatypeValue = item;
                        return;
                    }
                })
                this.objectOptions.forEach(function(item) {
                    if (item.value == vm.tableData[value].object) {
                        vm.objectValue = item;
                        return;
                    }
                })
                this.emailListOptions.forEach(function(item) {
                    if (item.value == vm.tableData[value].emaillistid) {
                        vm.emailListValue = item;
                        return;
                    }
                })

                this.showModal01 = true;
                this.metric = this.tableData[value].metric;
                this.critical = this.tableData[value].critical;
                this.warning = this.tableData[value].warning;
                this.interval = this.tableData[value].interval;

                if (this.objectValue.value == "prometheus") {
                    this.isDisabled = true;
                    this.isDisabledDM = false;
                } else {
                    this.isDisabled = false;
                    this.isDisabledDM = true;
                    this.dmValue = {};
                }
            },
            buttonClickNew() {
                this.moduleValue = "";
                this.dmValue = "";
                this.spaceValue = "";
                this.metric = "";
                this.critical = "";
                this.warning = "";
                this.interval = "";
                this.statusValue = this.statusOptions[0];
                this.datatypeValue = this.datatypeOptions[0];
                this.objectValue = this.objectOptions[0];
                this.showModal02 = true;
                this.emailListValue = "";

                if (this.objectValue.value == "prometheus") {
                    this.isDisabled = true;
                    this.isDisabledDM = false;
                } else {
                    this.isDisabled = false;
                    this.isDisabledDM = true;
                    this.dmValue = {};
                }
            },
            selectRight() {
                let vm = this;
                vm.currentPage++;
                vm.putTaskSelect(vm.selectParam.next);
                vm.pages[vm.currentPage] = vm.selectParam.next;
            },
            selectLeft() {
                let vm = this;
                if (vm.currentPage > 0) {
                    vm.currentPage--;
                }
                vm.putTaskSelect(vm.pages[vm.currentPage]);
            },
            putTaskSelect(current) {
                let vm = this;
                let dataPut = {}

                if (current != "") {
                    dataPut.current = current;
                } else {
                    dataPut.current = "";
                }
                dataPut.pagesize = 100;

                vm.tableLoading = true;
                $.ajax({
                    url: "/api/v1/admin/task/select?" + Math.random(),
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    success: function (data) {
                        if (data.data == null) {
                            data.data = [];
                        }
                        vm.tableData = data.data;
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
                    }
                });
            },
            putTaskUpdate() {
                let vm = this;
                let dataPut = {
                    moduleid: this.moduleValue.value,
                    dmid: this.dmValue.value,
                    spaceid: this.spaceValue.value,
                    object: this.objectValue.value,
                    metric: this.metric,
                    status: parseInt(this.statusValue.value),
                    critical: this.critical,
                    warning: this.warning,
                    interval: parseInt(this.interval),
                    datatype: this.datatypeValue.value,
                    emaillistid: this.emailListValue.value,
                };

                $.ajax({
                    url: "/api/v1/admin/task/update",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('message.message01'), color: 'primary' });
                            vm.putTaskSelect();
                            vm.showModal01 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('message.message02'), color: 'danger' });
                            vm.showModal01 = true;
                            return true;
                        },
                    }
                });
            },
            rowClickRemove(value) {
                this.showModal03 = true;
                this.rowIndexRemove = value;
                this.taskModulIDRemove = this.tableData[value].moduleid;
                this.taskSpaceIDRemove = this.tableData[value].spaceid;
                this.taskObjectRemove = this.tableData[value].object;
                this.taskMetricRemove = this.tableData[value].metric;
            },
            putTaskRemove(value) {
                let vm = this;
                let dataPut = {
                    moduleid: this.tableData[value].moduleid,
                    spaceid: this.tableData[value].spaceid,
                    object: this.tableData[value].object,
                    metric: this.tableData[value].metric,
                };
                $.ajax({
                    url: "/api/v1/admin/task/remove",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('task.message17'), color: 'primary' });
                            vm.putTaskSelect();
                            vm.showModal03 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('task.message18'), color: 'danger' });
                            vm.showModal03 = true;
                            return true;
                        },
                    }
                });
            },
            putTaskCreate() {
                let vm = this;

                if (this.status == "") {
                    this.status = 1;
                }
                if (this.moduleValue.value == "" || this.moduleValue.value == undefined) {
                    this.showModal02 = true;
                    this.isError01 = true;
                    this.$vaToast.init({ message: this.$t('task.message02'), color: 'danger' });
                    return;
                }
                if (this.spaceValue.value == "" || this.spaceValue.value == undefined) {
                    this.showModal02 = true;
                    this.isError02 = true;
                    this.$vaToast.init({ message: this.$t('task.message01'), color: 'danger' });
                    return;
                }
                if (this.objectValue.value == "" || this.objectValue.value == undefined) {
                    this.showModal02 = true;
                    this.isError03 = true;
                    this.$vaToast.init({ message: this.$t('task.message03'), color: 'danger' });
                    return;
                }
                if (this.metric == "") {
                    this.showModal02 = true;
                    this.isError04 = true;
                    this.$vaToast.init({ message: this.$t('task.message05'), color: 'danger' });
                    return;
                }

                let dataPut = {
                    moduleid: this.moduleValue.value,
                    dmid: this.dmValue.value,
                    spaceid: this.spaceValue.value,
                    object: this.objectValue.value,
                    metric: this.metric,
                    status: parseInt(this.statusValue.value),
                    critical: this.critical,
                    warning: this.warning,
                    interval: parseInt(this.interval),
                    datatype: this.datatypeValue.value,
                    emaillistid: this.emailListValue.value,
                };
                $.ajax({
                    url: "/api/v1/admin/task/create",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('task.message19'), color: 'primary' });
                            vm.putTaskSelect();
                            vm.showModal02 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('task.message20'), color: 'danger' });
                            vm.showModal02 = true;
                            return true;
                        },
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
                        if (data.role == "superadmin" || data.role == "admin") {
                            vm.userBlock = false
                        } else {
                            vm.userBlock = true
                        }
                        return true;
                    }
                });

                this.colDefs = [
                    { field: "moduledesc", headerName: this.$t('message.module') },
                    { field: "spacedesc", headerName: this.$t('message.space') },
                    { field: "object", headerName: this.$t('task.object') },
                    { field: "metric", headerName: this.$t('task.metric') },
                    { field: "interval", headerName: this.$t('task.interval') },
                    { field: "status", headerName: this.$t('message.status') },
                    { 
                        field: "action",
                        headerName: this.$t('message.action'),
                        cellRenderer: 'actionRenderer',
                        cellRendererParams: {
                            rowClickEdit: this.rowClickEdit,
                            rowClickRemove: this.rowClickRemove,
                            userBlock: this.userBlock,
                        },
                    },
                ];

                this.putTaskSelect();
                this.spaceSelect();
                this.moduleSelect();
                this.emailListSelect();
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
            moduleSelect() {
                let vm = this;

                let dataPut = {
                    pagesize: 0,
                };

                $.ajax({
                    url: "/api/v1/admin/registration/select?" + Math.random(),
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    success: function (data) {
                        if (data.data == null) {
                            data.data = [];
                        } else {
                            if (data.data.length > 0) {
                                let iDM = 0;
                                data.data.forEach(function(item, i) {
                                    vm.moduleOptions[i] = { text: item.description, value: item.id };
                                    if (item.type == "datamanager") {
                                        vm.dmOptions[iDM] = { text: item.description, value: item.id };
                                        iDM++;
                                    }
                                });
                            }
                        }
                        return true;
                    }
                });
            },
            emailListSelect() {
                let vm = this;
                $.ajax({
                    url: "/api/v1/admin/emaillist/select?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        if (data == null) {
                            data = [];
                        } else {
                            if (data.length > 0) {
                                let iA = 0;
                                data.forEach(function(item) {
                                    if (item.status == 1) {
                                        vm.emailListOptions[iA] = { text: item.name, value: item.id };
                                        iA++;
                                    }
                                });
                            }
                        }
                        return true;
                    }
                });
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
                if (newValue != oldValue && oldValue != undefined) {
                    this.showModal01 = false;
                    this.showModal02 = false;
                    this.colDefs = [
                        { field: "moduledesc", headerName: this.$t('message.module') },
                        { field: "spacedesc", headerName: this.$t('message.space') },
                        { field: "object", headerName: this.$t('task.object') },
                        { field: "metric", headerName: this.$t('task.metric') },
                        { field: "interval", headerName: this.$t('task.interval') },
                        { field: "status", headerName: this.$t('message.status') },
                        { 
                            field: "action",
                            headerName: this.$t('message.action'),
                            cellRenderer: 'actionRenderer',
                            cellRendererParams: {
                                rowClickEdit: this.rowClickEdit,
                                rowClickRemove: this.rowClickRemove,
                                userBlock: this.userBlock,
                            },
                        },
                    ];
                    this.statusOptions = [
                        {text: this.$t('message.on'), value: 1},
                        {text: this.$t('message.off'), value: 0},
                    ];
                    this.datatypeOptions = [
                        {text: this.$t('message.number'), value: "float"},
                        {text: this.$t('message.text'), value: "text"},
                    ];
                    this.componentKey += 1;
                }
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