<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.spaces')" to="/space" disabled />
        </va-breadcrumbs>
    </div>

    <div class="row layout va-gutter-4 md12 justify-end">
        <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickNew" class="mr-3" />
    </div>
    
    <div class="layout va-gutter-4 md12">
        <va-inner-loading :loading="tableLoading">
            <ag-grid-vue :key="componentKey" :rowData="tableData" :columnDefs="colDefs" domLayout="autoHeight" :components="components" @grid-ready="onGridReady" style="height: 100%; width: 100%"></ag-grid-vue>
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
            spaceSelect() {
                const vm = this;
                vm.tableLoading = true;
                $.ajax({
                    url: "/api/v1/admin/space/select?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        if (data == null) {
                            data = [];
                        }
                        vm.tableData = data;
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
                            vm.spaceSelect();
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
                            vm.spaceSelect();
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
                            vm.$vaToast.init({ message: vm.$t('space.message06'), color: 'primary' });
                            vm.spaceSelect();
                            vm.showModal02 = false;
                            return;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('space.message07'), color: 'danger' });
                            vm.showModal02 = true;
                            return;
                        },
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
                            vm.userBlock = false
                        } else {
                            vm.userBlock = true
                        }
                        return true;
                    }
                });
                this.spaceSelect();
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