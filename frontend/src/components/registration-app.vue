<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.modules')" to="/registration" disabled />
        </va-breadcrumbs>
    </div>

    <div class="row layout va-gutter-3">
        <div class="flex layout va-gutter-3 md6" align="left">
            <va-button icon="chevron_left" v-on:click="selectLeft()" v-bind:disabled="isDisabledLeft" class="mr-2 ml-1" />
            <va-button disabled>{{ currentPage + 1 }}</va-button>
            <va-button icon="chevron_right" v-on:click="selectRight()" v-bind:disabled="isDisabledRight"  class="ml-2" />
        </div>
        <div class="flex layout va-gutter-3 md6" align="right">
            <va-button icon="refresh" v-on:click="putRegistrationSelect(pages[currentPage])" class="ml-2 mr-1" />
            <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickNew" class="ml-2 mr-1" />
        </div>
    </div>

    <div class="layout va-gutter-4 md12">
        <va-inner-loading :loading="tableLoading">
            <ag-grid-vue :key="componentKey" :rowData="tableData" :columnDefs="colDefs" domLayout="autoHeight" :components="components" @grid-ready="onGridReady" style="height: 100%; width: 100%" :enableCellTextSelection="true" :ensureDomOrder="true"></ag-grid-vue>
        </va-inner-loading>
    </div>
    
    <va-modal v-model="showModal02" :title="$t('module.new')" v-on:ok="putRegistrationCreate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-select v-model="typeValue" class="mb-4" :placeholder="$t('module.message11')" :label="$t('message.type')" v-model:options="typeOptions" @update:modelValue="typeUpdate()" max-height="150px"></va-select></div>
        <div><va-input class="mb-4" v-model="address" :placeholder="$t('module.message13')" :label="$t('message.address')" v-bind:disabled="isDisabled" v-bind:error="isError02"></va-input></div>
        <div><va-select v-model="statusValue" class="mb-4" :placeholder="$t('module.message09')" :label="$t('message.status')" v-model:options="statusOptions" max-height="150px"></va-select></div>
        <div><va-input class="mb-4" v-model="description" :placeholder="$t('module.message10')" :label="$t('message.description')" :max-length="60" counter :rules="[(v) => v.length < 60]" v-bind:error="isError01"></va-input></div>
    </va-modal>
    <va-modal v-model="showModal01" :title="$t('module.message01')" v-on:ok="putRegistrationUpdate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-input class="mb-4" v-model="id" :label="$t('module.id')" disabled></va-input></div>
        <div><va-select v-model="typeValue" class="mb-4" :placeholder="$t('module.message11')" :label="$t('message.type')" v-model:options="typeOptions" @update:modelValue="typeUpdate()" max-height="150px"></va-select></div>
        <div><va-input class="mb-4" v-model="address" :placeholder="$t('module.message13')" :label="$t('message.address')" v-bind:disabled="isDisabled" v-bind:error="isError02"></va-input></div>
        <div><va-select v-model="statusValue" class="mb-4":placeholder="$t('module.message09')" :label="$t('message.status')" v-model:options="statusOptions" max-height="150px"></va-select></div>
        <div><va-input class="mb-4" v-model="description" :placeholder="$t('module.message10')" :label="$t('message.description')" :max-length="60" counter :rules="[(v) => v.length < 60]" v-bind:error="isError01"></va-input></div>
    </va-modal>
    <va-modal v-model="showModal03" :title="$t('module.message12')" v-on:ok="putRegistrationRemove(rowIndexRemove)" :okText="$t('message.delete')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div>
            {{ $t('module.message08') }} {{ registrationIDRemove }}<br><br>
            <b>{{ $t('module.message15') }}</b>
        </div>
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
            const typeOptions = [
                {text: this.$t('message.notdefined'), value: ""},
                {text: "zbxagent", value: "zbxagent"},
                {text: "prometheus", value: "prometheus"},
                {text: "uimanager", value: "uimanager"},
                {text: "datamanager", value: "datamanager"},
            ];

            return {
                isError01: false,
                isError02: false,
                isDisabled: true,
                userRole: null,
                userBlock: true,
                rowIndexRemove: null,
                registrationIDRemove: "",
                noData: this.$t('message.nodata'),
                showModal01: false,
                showModal02: false,
                showModal03: false,
                tableData: [],
                id: "",
                statusOptions: statusOptions,
                statusValue: statusOptions[0],
                typeOptions: typeOptions,
                typeValue: typeOptions[0],
                address: "",
                description: "",
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
                    { field: "id", headerName: this.$t('module.id') },
                    { field: "type", headerName: this.$t('message.type') },
                    { field: "address", headerName: this.$t('message.address') },
                    { field: "status", headerName: this.$t('message.status') },
                    { field: "description", headerName: this.$t('message.description') },
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
            typeUpdate() {
                if (this.typeValue.value == "prometheus") {
                    this.isDisabled = false;
                } else {
                    this.isDisabled = true;
                }
            },
            rowClickEdit(value) {
                let vm = this;
                this.statusOptions.forEach(function(item) {
                        if (item.value == vm.tableData[value].status) {
                            vm.statusValue = item;
                            return;
                        }
                    }
                )
                this.typeOptions.forEach(function(item) {
                        if (item.value == vm.tableData[value].type) {
                            vm.typeValue = item;
                            return;
                        }
                    }
                )
                this.showModal01 = true;
                this.id = this.tableData[value].id;
                this.type = this.tableData[value].type;
                this.address = this.tableData[value].address;
                this.description = this.tableData[value].description;

                if (this.typeValue.value == "prometheus") {
                    this.isDisabled = false;
                } else {
                    this.isDisabled = true;
                }
            },
            buttonClickNew() {
                this.id = "";
                this.description = "";
                this.address = "";
                this.statusValue = this.statusOptions[0];
                this.typeValue = this.typeOptions[0];
                this.showModal02 = true;

                if (this.typeValue.value == "prometheus") {
                    this.isDisabled = false;
                } else {
                    this.isDisabled = true;
                }
            },
            selectRight() {
                let vm = this;
                vm.currentPage++;
                vm.putRegistrationSelect(vm.selectParam.next);
                vm.pages[vm.currentPage] = vm.selectParam.next;
            },
            selectLeft() {
                let vm = this;
                if (vm.currentPage > 0) {
                    vm.currentPage--;
                }
                vm.putRegistrationSelect(vm.pages[vm.currentPage]);
            },
            putRegistrationSelect(current) {
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
                    url: "/api/v1/admin/registration/select?" + Math.random(),
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
            putRegistrationUpdate() {
                let vm = this;

                if (this.typeValue.value == "prometheus" && (this.address == "" || this.address == undefined)) {
                    this.showModal01 = true;
                    this.isError02 = true;
                    this.$vaToast.init({ message: vm.$t('module.message14'), color: 'danger' });
                    return;
                }

                let dataPut = {
                    id: this.id,
                    type: this.typeValue.value,
                    address: this.address,
                    status: parseInt(this.statusValue.value),
                    description: this.description,
                };
                $.ajax({
                    url: "/api/v1/admin/registration/update",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('message.message01'), color: 'primary' });
                            vm.putRegistrationSelect();
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
                this.registrationIDRemove = this.tableData[value].id;
            },
            putRegistrationRemove(value) {
                let vm = this;
                let dataPut = {
                    id: this.tableData[value].id
                };
                $.ajax({
                    url: "/api/v1/admin/registration/remove",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('module.message02'), color: 'primary' });
                            vm.putRegistrationSelect();
                            vm.showModal03 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('module.message03'), color: 'danger' });
                            vm.showModal03 = true;
                            return true;
                        },
                    }
                });
            },
            putRegistrationCreate() {
                let vm = this;
                
                if (this.description == "" || this.description == undefined) {
                    this.showModal02 = true;
                    this.isError01 = true;
                    this.$vaToast.init({ message: vm.$t('module.message04'), color: 'danger' });
                    return;
                }
                if (this.description.length > 60) {
                    this.showModal02 = true;
                    this.isError01 = true;
                    this.$vaToast.init({ message: vm.$t('module.message05'), color: 'danger' });
                    return;
                }
                if (this.typeValue.value == "prometheus" && (this.address == "" || this.address == undefined)) {
                    this.showModal02 = true;
                    this.isError02 = true;
                    this.$vaToast.init({ message: vm.$t('module.message14'), color: 'danger' });
                    return;
                }

                let dataPut = {
                    type: this.typeValue.value,
                    address: this.address,
                    status: parseInt(this.statusValue.value),
                    description: this.description,
                };
                $.ajax({
                    url: "/api/v1/admin/registration/create",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('module.message06'), color: 'primary' });
                            vm.putRegistrationSelect();
                            vm.showModal02 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('module.message07'), color: 'danger' });
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
                    { field: "id", headerName: this.$t('module.id') },
                    { field: "type", headerName: this.$t('message.type') },
                    { field: "address", headerName: this.$t('message.address') },
                    { field: "status", headerName: this.$t('message.status') },
                    { field: "description", headerName: this.$t('message.description') },
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

                this.putRegistrationSelect();
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
                        { field: "id", headerName: this.$t('module.id') },
                        { field: "type", headerName: this.$t('message.type') },
                        { field: "address", headerName: this.$t('message.address') },
                        { field: "status", headerName: this.$t('message.status') },
                        { field: "description", headerName: this.$t('message.description') },
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
                    this.typeOptions = [
                        {text: this.$t('message.notdefined'), value: ""},
                        {text: "zbxagent", value: "zbxagent"},
                        {text: "prometheus", value: "prometheus"},
                        {text: "uimanager", value: "uimanager"},
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