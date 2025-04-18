<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.users')" to="/user" disabled />
        </va-breadcrumbs>
    </div>

    <div class="row layout va-gutter-3 md12 justify-end">
        <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickNew" class="mr-3" />
    </div>
    
    <div class="layout va-gutter-3 md12">
        <va-inner-loading :loading="tableLoading">
            <ag-grid-vue :key="componentKey" :rowData="tableData" :columnDefs="colDefs" domLayout="autoHeight" :components="components" @grid-ready="onGridReady" style="height: 100%; width: 100%"></ag-grid-vue>
        </va-inner-loading>
    </div>
    
    <va-modal v-model="showModal02" :title="$t('user.new')" v-on:ok="putUserCreate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-input class="mb-4" v-model="userName" :placeholder="$t('user.username')" :label="$t('user.username')"></va-input></div>
        <div><va-input type="password" class="mb-4" v-model="newPassword" :placeholder="$t('message.password')" :label="$t('message.password')"></va-input></div>
        <div><va-input class="mb-4" v-model="firstName" :placeholder="$t('user.firstname')" :label="$t('user.firstname')"></va-input></div>
        <div><va-input class="mb-4" v-model="lastName" :placeholder="$t('user.lastname')" :label="$t('user.lastname')"></va-input></div>
        <div><va-select v-model="roleValue" class="mb-4" :placeholder="$t('user.role')" :label="$t('user.role')" v-model:options="roleOptions" max-height="150px"></va-select></div>
        <div><va-select v-model="statusValue" class="mb-4" :placeholder="$t('message.status')" :label="$t('message.status')" v-model:options="statusOptions" max-height="150px"></va-select></div>
        <div><va-input class="mb-4" v-model="description" :placeholder="$t('message.description')" :label="$t('message.description')"></va-input></div>
    </va-modal>
    <va-modal v-model="showModal01" :title="$t('user.message08')" v-on:ok="putUserUpdate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div><va-input class="mb-4" v-model="userID" label="User ID" disabled></va-input></div>
        <div><va-input class="mb-4" v-model="userName" :label="$t('user.username')" disabled></va-input></div>
        <div><va-input type="password" class="mb-4" v-model="newPassword" :placeholder="$t('password.new')" :label="$t('password.new')"></va-input></div>
        <div><va-input class="mb-4" v-model="firstName" :placeholder="$t('user.firstname')" :label="$t('user.firstname')"></va-input></div>
        <div><va-input class="mb-4" v-model="lastName" :placeholder="$t('user.lastname')" :label="$t('user.lastname')"></va-input></div>
        <div><va-select v-model="roleValue" class="mb-4" :placeholder="$t('user.role')" :label="$t('user.role')" v-model:options="roleOptions" max-height="150px"></va-select></div>
        <div><va-select v-model="statusValue" class="mb-4" :placeholder="$t('message.status')" :label="$t('message.status')" v-model:options="statusOptions" max-height="150px"></va-select></div>
        <div><va-input class="mb-4" v-model="description" :placeholder="$t('message.description')" :label="$t('message.description')"></va-input></div>
    </va-modal>
    <va-modal v-model="showModal03" :title="$t('user.message07')" v-on:ok="putUserRemove(rowIndexRemove)" :okText="$t('message.delete')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div>{{ $t('user.message09') }} {{ userNameRemove }}</div>
    </va-modal>
</template>

<script>
    import $ from 'jquery'
    import sha256 from 'sha256'
    import moment from 'moment'
    import { mapState } from 'vuex';
    import { AgGridVue } from "ag-grid-vue3"; // Vue Data Grid Component
    import ActionRenderer from './ag-grid-app-action.vue'

    export default {
        components: {
            AgGridVue,
        },
        data() {
            const statusOptions = [
                { text: this.$t('message.on'), value: 1 },
                { text: this.$t('message.off'), value: 0 },
            ];
            const roleOptions = [
                { text: this.$t('role.user'), value: "user" },
                { text: this.$t('role.admin'), value: "admin" },                
                { text: this.$t('role.superadmin'), value: "superadmin" },
            ];

            return {
                userRole: null,
                userBlock: true,
                rowIndexRemove: null,
                userNameRemove: "",
                noData: this.$t('message.nodata'),
                showModal01: false,
                showModal02: false,
                showModal03: false,
                tableData: [],
                userID: "",
                userName: "",
                newPassword: "",
                firstName: "",
                lastName: "",
                statusOptions: statusOptions,
                statusValue: statusOptions[0],
                roleOptions: roleOptions,
                roleValue: roleOptions[0],
                description: "",
                tableLoading: false,
                gridApi: null,
                componentKey: 0,

                // Column Definitions: Defines the columns to be displayed.
                colDefs: [
                    { field: "userid", headerName: this.$t('user.userid') },
                    { field: "username", headerName: this.$t('user.username') },
                    { field: "firstname", headerName: this.$t('user.firstname') },
                    { field: "lastname", headerName: this.$t('user.lastname') },
                    { field: "role", headerName: this.$t('user.role') },
                    { field: "status", headerName: this.$t('message.status') },
                    { field: "logintime", headerName: this.$t('user.logintime') },
                    { field: "ownerid", headerName: this.$t('user.ownerid') },
                    { field: "description", headerName: this.$t('message.description') },
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
                }
            }
        },
        methods: {
            rowClickEdit(value) {
                let vm = this;
                if (this.userRole == "admin") {
                    this.roleOptions = [
                        { text: this.$t('role.user'), value: "user" },
                    ];
                } else {
                    this.roleOptions = [
                        { text: this.$t('role.user'), value: "user" },
                        { text: this.$t('role.admin'), value: "admin" },                
                        { text: this.$t('role.superadmin'), value: "superadmin" },
                    ];
                }
                this.statusOptions.forEach(function(item) {
                        if (item.value == vm.tableData[value].status) {
                            vm.statusValue = item;
                            return;
                        }
                    }
                )
                this.roleOptions.forEach(function(item) {
                        if (item.value == vm.tableData[value].role) {
                            vm.roleValue = item;
                            return;
                        }
                    }
                )
                this.showModal01 = true;
                this.userID = this.tableData[value].userid;
                this.userName = this.tableData[value].username;
                this.newPassword = "";
                this.firstName = this.tableData[value].firstname;
                this.lastName = this.tableData[value].lastname;
                this.description = this.tableData[value].description;
            },
            buttonClickNew() {
                this.userName = "";
                this.newPassword = "";
                this.firstName = "";
                this.lastName = "";
                this.description = "";
                this.statusValue = this.statusOptions[0];
                this.showModal02 = true;
                if (this.userRole == "admin") {
                    this.roleOptions = [
                        { text: this.$t('role.user'), value: "user" },
                    ];
                } else {
                    this.roleOptions = [
                        { text: this.$t('role.user'), value: "user" },
                        { text: this.$t('role.admin'), value: "admin" },                
                        { text: this.$t('role.superadmin'), value: "superadmin" },
                    ];
                }
                this.roleValue = this.roleOptions[0];
            },
            userSelect() {
                const vm = this;
                vm.tableLoading = true;
                $.ajax({
                    url: "/api/v1/admin/user/select?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        let dataNormal = [];
                        if (data != null) {
                            data.forEach(function(item, i) {
                                dataNormal[i] = item;
                                dataNormal[i].logintime = moment(item.logintime).format("DD.MM.YYYY HH:mm:ss");
                            });
                        }
                        
                        vm.tableData = dataNormal;
                        vm.tableLoading = false;
                        window.onresize = () => { vm.wrapperSize = document.documentElement.clientHeight - 110 };
                        vm.componentKey += 1; // Run the component DOM interporeing for the correct calculation of scrolling
                        return true;
                    }
                });
            },
            putUserUpdate() {
                let vm = this;
                let sha256Password = "";
                if (this.newPassword != "") {
                    sha256Password = sha256(this.newPassword);
                }
                let dataPut = {
                    username: this.userName,
                    password: sha256Password,
                    firstname: this.firstName,
                    lastname: this.lastName,
                    status: parseInt(this.statusValue.value),
                    description: this.description,
                    role: this.roleValue.value,
                };
                $.ajax({
                    url: "/api/v1/admin/user/update",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('message.message01'), color: 'primary' });
                            vm.userSelect();
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
                this.rowIndexRemove = value;
                this.userNameRemove = this.tableData[value].username;
                this.showModal03 = true;
            },
            putUserRemove(value) {
                let vm = this;
                let dataPut = {
                    username: this.tableData[value].username,
                    userid: this.tableData[value].userid,
                    ownerid: this.tableData[value].ownerid,
                };
                $.ajax({
                    url: "/api/v1/admin/user/remove",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('user.message01'), color: 'primary' });
                            vm.userSelect();
                            vm.showModal03 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('user.message02'), color: 'danger' });
                            vm.showModal03 = true;
                            return true;
                        },
                    }
                });
            },
            putUserCreate() {
                let vm = this;
                if (this.userName == "") {
                    vm.showModal02 = true;
                    vm.$vaToast.init({ message: vm.$t('user.message03'), color: 'warning' });
                    return;
                }
                if (this.newPassword == "") {
                    vm.showModal02 = true;
                    vm.$vaToast.init({ message: vm.$t('user.message04'), color: 'warning' });
                    return;
                }
                let dataPut = {
                    username: this.userName,
                    password: sha256(this.newPassword),
                    firstname: this.firstName,
                    lastname: this.lastName,
                    status: parseInt(this.statusValue.value),
                    description: this.description,
                    role: this.roleValue.value,
                };
                $.ajax({
                    url: "/api/v1/admin/user/create",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('user.message05'), color: 'primary' });
                            vm.userSelect();
                            vm.showModal02 = false;
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('user.message06'), color: 'danger' });
                            vm.showModal02 = true;
                            return true;
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
                            vm.userData = vm.$t('message.noaccess')
                        }
                        return true;
                    }
                });
                this.userSelect();
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
        created() {
            this.userInfo();
            window.onresize = () => { this.wrapperSize = document.documentElement.clientHeight - 110 };
        },
        computed: {
            ...mapState(['locale'])
        },
        watch: {
            locale(newValue, oldValue) {
                if (newValue != oldValue && oldValue != undefined) {
                    this.showModal01 = false;
                    this.showModal02 = false;
                    this.colDefs = [
                        { field: "userid", headerName: this.$t('user.userid') },
                        { field: "username", headerName: this.$t('user.username') },
                        { field: "firstname", headerName: this.$t('user.firstname') },
                        { field: "lastname", headerName: this.$t('user.lastname') },
                        { field: "role", headerName: this.$t('user.role') },
                        { field: "status", headerName: this.$t('message.status') },
                        { field: "logintime", headerName: this.$t('user.logintime') },
                        { field: "ownerid", headerName: this.$t('user.ownerid') },
                        { field: "description", headerName: this.$t('message.description') },
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
                        { text: this.$t('message.on'), value: 1 },
                        { text: this.$t('message.off'), value: 0 },
                    ];
                    this.roleOptions = [
                        { text: this.$t('role.user'), value: "user" },
                        { text: this.$t('role.admin'), value: "admin" },                
                        { text: this.$t('role.superadmin'), value: "superadmin" },
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