<template>
    <div class="bredcrumbs">
        <va-breadcrumbs class="mb-4">
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.settings')" to="/settings" disabled />
        </va-breadcrumbs>
    </div>

    <va-modal v-model="showModalCreate" :title="$t('emaillist.message07')" v-on:ok="putEMailListCreate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div class="row">
            <div class="flex flex-col md12"><va-input class="mb-2" v-model="emailListName" :placeholder="$t('emaillist.message01')" :label="$t('emaillist.message02')" /></div>
            <div class="flex flex-col md12"><va-textarea class="mb-2" v-model="emailListDescription" :placeholder="$t('emaillist.message03')" :label="$t('emaillist.message04')" /></div>
        </div>
    </va-modal>

    <VaLayout>
        <template #left>
            <div style="display: flex; height: 100%;">
            <VaSidebar v-model="showSidebar"
                :color="cardColor01"
                :active-color="sidebarAColor"
                text-color="textPrimary"
            >
                <VaSidebarItem>
                    <VaSidebarItemContent>
                        <VaSidebarItemTitle class="whitespace-normal">{{ $t('message.settings') }}</VaSidebarItemTitle>
                    </VaSidebarItemContent>
                </VaSidebarItem>
                <VaSidebarItem :active="page === 1" @click="page = 1">
                    <VaSidebarItemContent>
                        <VaIcon name="email" /> 
                        <VaSidebarItemTitle>{{ $t('emaillist.message05') }}</VaSidebarItemTitle>
                    </VaSidebarItemContent>
                </VaSidebarItem>
            </VaSidebar>
            </div>
        </template>
        <template #content>
            <main v-if="page === 1" class="layout va-gutter-3">
                <div class="row layout va-gutter-3">
                    <div class="flex layout va-gutter-3 md12 va-h6" align="left">
                        <b>{{ $t('emaillist.message06') }}</b>
                    </div>
                </div>
                <div class="row layout va-gutter-3">
                    <div class="flex layout va-gutter-3 md6" align="left">
                        <VaButton :icon="showSidebar ? 'menu_open' : 'menu'" @click="showSidebar = !showSidebar" />
                    </div>
                    <div class="flex layout va-gutter-3 md6" align="right">
                        <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickNew" />
                    </div>
                </div>
                <va-inner-loading :loading="innerLoading">
                    <div class="row layout va-gutter-3">
                        <div class="flex layout va-gutter-3 md6" v-for="(emailList, index01) in emailLists" :key="index01">
                            <va-modal v-model="showModalRemove[index01]" :title="$t('emaillist.message08')" v-on:ok="putEMailListRemove(index01)" :okText="$t('message.delete')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
                                <div>{{ $t('emaillist.message10') }} {{ index01 + 1 }} ({{ emailList.name }})</div>
                            </va-modal>
                            <va-modal v-model="showModalEdit[index01]" :title="$t('emaillist.message09')" v-on:ok="putEMailListUpdate(index01)" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
                                <div class="row">
                                    <div class="flex flex-col md12"><va-input class="mb-2" v-model="emailListEditName" :placeholder="$t('emaillist.message01')" :label="$t('emaillist.message02')" /></div>
                                    <div class="flex flex-col md12"><va-textarea class="mb-2" v-model="emailListEditDescription" :placeholder="$t('emaillist.message03')" :label="$t('emaillist.message04')" /></div>
                                    <div class="flex flex-col md12"><va-select v-model="emailListEditStatusValue" class="mb-2" :placeholder="$t('emaillist.message11')" :label="$t('message.status')" v-model:options="emailListEditStatusOptions"></va-select></div>
                                </div>
                                <div class="md12">
                                    <va-divider dashed>
                                        <span class="px-2">{{ $t('emaillist.message12') }}</span>
                                    </va-divider>
                                </div>
                                <div class="md12" align="center">
                                    <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickDataNew(index01)" class="mr-3" />
                                </div>
                                <div class="flex layout md12 va-gutter-3" v-for="(emailListData, index02) in emailListsEdit[index01].data" :key="index02">
                                    <va-card stripe :stripe-color="stripeColorData[index02]">
                                        <va-card-title>{{ $t('emaillist.message13') }} {{ index02 + 1 }}</va-card-title>
                                        <va-card-content>
                                            <div class="flex flex-col md12"><va-input class="mb-2" v-model="emailListsEdit[index01].data[index02].email" :placeholder="$t('emaillist.message22')" :label="$t('emaillist.message13')" /></div>
                                            <div class="flex flex-col md12"><va-input class="mb-2" v-model="emailListsEdit[index01].data[index02].description" :placeholder="$t('emaillist.message24')" :label="$t('emaillist.message23')" /></div>
                                            <div class="flex flex-col md12"><va-select class="mb-2" v-model="emailListDataEditStatusValue[index01][index02]" :options="emailListDataEditStatusOptions" :placeholder="$t('emaillist.message11')" :label="$t('message.status')" :no-options-text="$t('message.listempty')"></va-select></div>
                                            <div class="md12" align="center">
                                                <va-button icon="delete" v-bind:disabled="userBlock" v-on:click="buttonClickDataRemove(index01, index02)" class="mr-3" />
                                            </div>
                                        </va-card-content>
                                    </va-card>
                                </div>
                                <div class="md12">
                                    <va-divider dashed />
                                </div>
                            </va-modal>
                            <va-card :color="cardColor01" stripe :stripe-color="stripeColor[index01]">
                                <va-card-title>
                                    <div class="layout va-gutter-1 md6 align-content-center justify-start">{{ index01 + 1 }} {{ emailList.name }}</div>
                                    <div class="row layout va-gutter-1 va-spacing-x-1 md6 justify-end">
                                        <va-button icon="edit" round v-on:click="putEMailListUpdateForm(index01)" />
                                        <va-button icon="delete" color="danger" round v-on:click="putEMailListRemoveConfirm(index01)" />
                                    </div>
                                </va-card-title>
                                <va-card-content>
                                    {{ emailList.description }}
                                </va-card-content>
                            </va-card>
                        
                        </div>
                    </div>
                </va-inner-loading>
            </main>
        </template>
    </VaLayout>
</template>

<script>
    import $ from 'jquery'
    import { mapState } from 'vuex'
    import { isValidEMail } from '../chrds.js'

    export default {
        name: 'settings-app',
        data() {
            const emailListEditStatusOptions = [
                {text: this.$t('message.on'), value: 1},
                {text: this.$t('message.off'), value: 0}
            ];

            return {
                cardColorL01: "#ade2ffff",
                cardColorD01: "#36536bff",
                sidebarAColorL: '#70b8e0ff',
                sidebarAColorD: '#1f303eff',
                cardColor01: "",
                sidebarAColor: "",

                showSidebar: true,
                page: 1,

                emailListName: "",
                emailListDescription: "",

                emailListEditName: "",
                emailListEditDescription: "",

                emailListEditStatusOptions: emailListEditStatusOptions,
                emailListEditStatusValue: emailListEditStatusOptions[0],
                emailListDataEditStatusOptions: emailListEditStatusOptions,
                emailListDataEditStatusValue: [],

                emailLists: [],
                emailListsEdit: [],

                stripeColor: [],
                stripeColorData: [],

                /* Переменные состояния модальных окон */
                showModalCreate: false,
                showModalRemove: [],
                showModalEdit: [],
            }
        },
        methods: {
            buttonClickNew() {
                this.emailListName = "";
                this.emailListDescription = "";

                this.showModalCreate = true;
            },
            putEMailListCreate() {
                const vm = this;

                if (vm.emailListName == "") {
                    vm.showModalCreate = true;
                    vm.$vaToast.init({ message: vm.$t('emaillist.message14'), color: 'danger' });
                    return;
                }

                let dataPut = {
                    name: vm.emailListName,
                    description: vm.emailListDescription,
                    status: 1
                };
                $.ajax({
                    url: "/api/v1/admin/emaillist/create",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('emaillist.message15'), color: 'primary' });
                            vm.showModalCreate = false;
                            vm.innerLoading = true;
                            vm.barLoading = false;
                            vm.getEMailListSelect();
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('emaillist.message16'), color: 'danger' });
                            vm.showModalCreate = true;
                            return true;
                        },
                    }
                });
            },
            getEMailListSelect() {
                const vm = this;
                $.ajax({
                    url: "/api/v1/admin/emaillist/select?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        if (data == null) {
                            data = [];
                            vm.emailLists = [];
                            vm.innerLoading = false;
                            return
                        } 
                        data.sort(function (a, b) {
                            if (a.name > b.name) {
                                return 1;
                            }
                            if (a.name < b.name) {
                                return -1;
                            }
                            return 0;
                        });
                        vm.emailLists = data;
                        data.forEach(function(item, i) {
                            vm.showModalRemove[i] = false;
                            vm.showModalEdit[i] = false;

                            if (item.status == 1) {
                                if (item.data == undefined) {
                                    vm.stripeColor[i] = "warning";
                                } else {
                                    vm.stripeColor[i] = "success";
                                }
                            } else {
                                vm.stripeColor[i] = "danger";
                            }
                        });
                    },
                });
            },
            buttonClickDataNew(index) {
                let size = 0;
                if (this.emailListsEdit[index].data == null) {
                    this.emailListsEdit[index].data = [];
                    size = this.emailListsEdit[index].data.push(
                        {
                            id: "",
                            email: "",
                            description: "",
                            status: 1
                        }
                    );

                    let data01 = [];
                    data01[0] = "";
                    this.emailListDataEditStatusValue[index] = data01;
                } else {
                    size = this.emailListsEdit[index].data.push(
                        {
                            id: "",
                            email: "",
                            description: "",
                            status: 1
                        }
                    );
                }
                this.emailListDataEditStatusValue[index][size -1] = this.emailListDataEditStatusOptions[0];
                this.stripeColorData[size - 1] = "warning";
            },
            putEMailListUpdateForm(index) {
                let vm = this;

                this.emailListsEdit = JSON.parse(JSON.stringify(this.emailLists));
                console.log(this.emailListsEdit);

                this.emailListEditName = this.emailListsEdit[index].name;
                this.emailListEditDescription = this.emailListsEdit[index].description;
                this.emailListEditStatus = this.emailListsEdit[index].status;

                this.emailListEditStatusOptions.forEach(function(item) {
                    if (item.value == vm.emailListsEdit[index].status) {
                        vm.emailListEditStatusValue = item;
                        return;
                    }
                })

                if (vm.emailListsEdit[index].data != null) {
                    let data01 = [];
                    vm.emailListsEdit[index].data.forEach(function(item, i) {
                        data01[i] = "";
                    });
                    vm.emailListDataEditStatusValue[index] = data01;
                }

                if (this.emailLists[index].data != null) {
                    this.emailLists[index].data.forEach(function(item, i) {
                        vm.stripeColorData[i] = "success";

                        vm.emailListEditStatusOptions.forEach(function(item) {
                            if (item.value == vm.emailListsEdit[index].data[i].status) {
                                vm.emailListDataEditStatusValue[index][i] = item;
                                return
                            }
                        });
                    });
                }

                this.showModalEdit[index] = true;
            },
            buttonClickDataRemove(index01, index02) {
                this.emailListsEdit[index01].data.splice(index02, 1);

                if (this.emailListsEdit[index01].data != null) {
                    this.emailListsEdit[index01].data.forEach(function(item, i) {
                        // Переопределение выбора значений для полей SELECT после сдвига массива
                        vm.emailListEditStatusOptions.forEach(function(item) {
                            if (item.value == vm.emailListsEdit[index01].data[i].status) {
                                vm.emailListDataEditStatusValue[index01][i] = item;
                                return
                            }
                        });
                    });
                }
            },
            putEMailListUpdate(index) {
                let vm = this;

                let errorFlag = 0;
                if (vm.emailListsEdit[index].data != undefined) {
                    vm.emailListsEdit[index].data.forEach(function(item, i) {
                        vm.emailListsEdit[index].data[i].status = vm.emailListDataEditStatusValue[index][i].value;
                        vm.emailListsEdit[index].data[i].email = vm.emailListsEdit[index].data[i].email;
                        vm.emailListsEdit[index].data[i].description = vm.emailListsEdit[index].data[i].description;

                        if (vm.emailListsEdit[index].data[i].email == "") {
                            vm.showModalEdit[index] = true;
                            vm.$vaToast.init({ message: vm.$t('emaillist.message17') + ' ' + (i + 1), color: 'warning' });
                            errorFlag = 1;
                            return;
                        }
                        if (isValidEMail(vm.emailListsEdit[index].data[i].email) == false) {
                            vm.showModalEdit[index] = true;
                            vm.$vaToast.init({ message: vm.$t('emaillist.message25') + ' ' + (i + 1), color: 'warning' });
                            errorFlag = 1;
                            return;
                        }
                    });
                }

                if (errorFlag == 1) {
                    return;
                }

                let dataPut = {
                    id: vm.emailListsEdit[index].id,
                    name: vm.emailListEditName,
                    description: vm.emailListEditDescription,
                    status: parseInt(vm.emailListEditStatusValue.value),
                    data: vm.emailListsEdit[index].data
                };

                $.ajax({
                    url: "/api/v1/admin/emaillist/edit",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('emaillist.message18'), color: 'primary' });
                            vm.showModalEdit[index] = false;
                            vm.innerLoading = true;
                            vm.barLoading = false;
                            vm.getEMailListSelect();
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('emaillist.message19'), color: 'danger' });
                            vm.showModalEdit[index] = true;
                            return true;
                        },
                    }
                });
            },
            putEMailListRemove(index) {
                let vm = this;

                $.ajax({
                    url: "/api/v1/admin/emaillist/remove",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(vm.emailLists[index]),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('emaillist.message20'), color: 'primary' });
                            vm.showModalRemove[index] = false;
                            vm.innerLoading = true;
                            vm.barLoading = false;
                            vm.getEMailListSelect();
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('emaillist.message21'), color: 'danger' });
                            vm.showModalRemove[index] = true;
                            return true;
                        },
                    }
                });
            },
            putEMailListRemoveConfirm(index) {
                this.showModalRemove[index] = true;
            },
        },
        created() {
            if (this.$store.state.mode == false) {
                this.cardColor01 = this.cardColorL01;
                this.sidebarAColor = this.sidebarAColorL;
            } else {
                this.cardColor01 = this.cardColorD01;
                this.sidebarAColor = this.sidebarAColorD;
            }

            this.getEMailListSelect();
        },
        unmounted() {
            clearInterval(this.chartDataIntervalID);
        },
        computed: {
            ...mapState(['locale', 'mode'])
        },
        watch: {
            locale(newValue, oldValue) {
                // Close all modal windows when changing the tongue
                if (newValue != oldValue && oldValue != undefined) {
                    for (let i = 0; i < this.showModalRemove.length; i++) {
                        this.showModalRemove[i] = false;
                    }
                    for (let i = 0; i < this.showModalEdit.length; i++) {
                        this.showModalEdit[i] = false;
                    }
                    this.showModalCreate = false;
                }
            },
            mode(newValue, oldValue) {
                if (newValue == false) {
                    this.cardColor01 = this.cardColorL01;
                    this.sidebarAColor = this.sidebarAColorL;

                } else {
                    this.cardColor01 = this.cardColorD01;
                    this.sidebarAColor = this.sidebarAColorD;
                }
            }
        }
    }
</script>

<style>
    div.bredcrumbs {
        margin: 1.5rem;
    }
    div.page {
        margin: 1.5rem;
    }
</style>