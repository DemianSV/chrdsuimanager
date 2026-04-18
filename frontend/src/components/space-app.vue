<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.spaces')" to="/space" disabled />
        </va-breadcrumbs>
    </div>

    <div class="row layout va-gutter-3">
        <div class="flex layout va-gutter-3 md6" align="left">
            <va-button icon="chevron_left" v-on:click="selectLeft()" v-bind:disabled="isDisabledLeft" class="mr-2 ml-1" />
            <va-button disabled>{{ currentPage + 1 }}</va-button>
            <va-button icon="chevron_right" v-on:click="selectRight()" v-bind:disabled="isDisabledRight"  class="ml-2" />
        </div>
        <div class="flex layout va-gutter-3 md6" align="right">
            <va-button icon="refresh" v-on:click="putSpaceSelect(pages[currentPage])" class="ml-2 mr-1" />
            <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickNew" class="ml-2 mr-1" />
        </div>
    </div>
    
    <div class="layout va-gutter-4 md12">
        <va-inner-loading :loading="tableLoading">
            <ag-grid-vue :key="componentKey" :rowData="tableData" :columnDefs="colDefs" domLayout="autoHeight" :components="components" @grid-ready="onGridReady" style="height: 100%; width: 100%" :enableCellTextSelection="true" :ensureDomOrder="true"></ag-grid-vue>
        </va-inner-loading>
    </div>
    
    <va-modal v-model="showModal02" :title="$t('space.new')" v-on:ok="putSpaceCreate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-input class="mb-4" v-model="description" :placeholder="$t('space.message01')" :label="$t('message.description')"></va-input></div>
        <div><va-select v-model="statusValue" class="mb-4" :placeholder="$t('space.message02')" :label="$t('message.status')" v-model:options="statusOptions" max-height="150px"></va-select></div>
    </va-modal>
    <va-modal v-model="showModal01" :title="$t('space.message03')" v-on:ok="putSpaceUpdate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-input class="mb-4" v-model="id" label="ID Пространства" disabled></va-input></div>
        <div><va-input class="mb-4" v-model="description" :placeholder="$t('space.message01')" :label="$t('message.description')"></va-input></div>
        <div><va-select v-model="statusValue" class="mb-4" :placeholder="$t('space.message02')" :label="$t('message.status')" v-model:options="statusOptions" max-height="150px"></va-select></div>
    </va-modal>
    <va-modal v-model="showModal03" :title="$t('space.message04')" v-on:ok="putSpaceRemove(rowIndexRemove)" :okText="$t('message.delete')":cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div>{{ $t('space.message09') }} {{ spaceIDRemove }}</div>
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

            return {
                userRole: null,
                userBlock: true,
                rowIndexRemove: null,
                spaceIDRemove: "",
                showModal01: false,
                showModal02: false,
                showModal03: false,
                tableData: [],
                id: "",
                spaceid: "",
                description: "",
                statusOptions: statusOptions,
                statusValue: statusOptions[0],
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
                    { field: "description", headerName: this.$t('message.description') },
                    { field: "status", headerName: this.$t('message.status') },
                    { field: "userid", headerName: this.$t('user.userid') },
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
            rowClickEdit(value) {
                let vm = this;
                this.statusOptions.forEach(function(item) {
                        if (item.value == vm.tableData[value].status) {
                            vm.statusValue = item;
                            return;
                        }
                    }
                )
                this.showModal01 = true;
                this.id = this.tableData[value].id;
                this.description = this.tableData[value].description;
            },
            buttonClickNew() {
                this.id = "";
                this.description = "";
                this.statusValue = this.statusOptions[0];
                this.showModal02 = true;
            },
            putSpaceSelect(current) {
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
                    url: "/api/v1/admin/space/select?" + Math.random(),
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
            putSpaceUpdate() {
                let vm = this;
                let dataPut = {
                    id: this.id,
                    description: this.description,
                    status: parseInt(this.statusValue.value),
                };
                $.ajax({
                    url: "/api/v1/admin/space/update",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('message.message01'), color: 'primary' });
                            vm.putSpaceSelect();
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
                this.spaceIDRemove = this.tableData[value].id;
            },
            putSpaceRemove(value) {
                let vm = this;
                let dataPut = {
                    id: this.tableData[value].id,
                };
                $.ajax({
                    url: "/api/v1/admin/space/remove",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('space.message05'), color: 'primary' });
                            vm.putSpaceSelect();
                            vm.showModal03 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('space.message06'), color: 'danger' });
                            vm.showModal03 = true;
                            return true;
                        },
                    }
                });
            },
            selectRight() {
                let vm = this;
                vm.currentPage++;
                vm.putSpaceSelect(vm.selectParam.next);
                vm.pages[vm.currentPage] = vm.selectParam.next;
            },
            selectLeft() {
                let vm = this;
                if (vm.currentPage > 0) {
                    vm.currentPage--;
                }
                vm.putSpaceSelect(vm.pages[vm.currentPage]);
            },
            putSpaceCreate() {
                let vm = this;
                let dataPut = {
                    id: this.moduleid,
                    description: this.description,
                    status: parseInt(this.statusValue.value),
                };
                $.ajax({
                    url: "/api/v1/admin/space/create",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('space.message07'), color: 'primary' });
                            vm.putSpaceSelect();
                            vm.showModal02 = false;
                            return;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('space.message08'), color: 'danger' });
                            vm.showModal02 = true;
                            return;
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
                        { field: "description", headerName: this.$t('message.description') },
                        { field: "status", headerName: this.$t('message.status') },
                        { field: "userid", headerName: this.$t('user.userid') },
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

                this.putSpaceSelect();
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
                        { field: "description", headerName: this.$t('message.description') },
                        { field: "status", headerName: this.$t('message.status') },
                        { field: "userid", headerName: this.$t('user.userid') },
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