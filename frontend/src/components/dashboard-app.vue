<template>
    <div class="row layout va-gutter-4 md12">
        <va-breadcrumbs>
            <va-breadcrumbs-item :label="$t('message.home')" to="/" />
            <va-breadcrumbs-item :label="$t('message.dashboards')" to="/dashboard" disabled />
        </va-breadcrumbs>
    </div>

    <div class="row layout va-gutter-3 md12 justify-end">
        <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickNew" class="mr-3" />
    </div>

    <va-modal v-model="showModalCreate" :title="$t('dashboard.message07')" v-on:ok="putDashboardCreate()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
        <div class="row">
            <div class="flex flex-col md12"><va-input class="mb-2" v-model="dashboardName" :placeholder="$t('dashboard.message01')" :label="$t('dashboard.message02')" /></div>
            <div class="flex flex-col md6"><va-date-input class="mb-2 pr-2" v-model="dashboardStartDate" :month-names="$tm('calendar.monthnames')" :weekday-names="$tm('calendar.weekdaynames')" :first-weekday="$t('calendar.firstweekdays')" :label="$t('dashboard.message03')" clearable /></div>
            <div class="flex flex-col md6"><va-time-input class="mb-2 pl-2" v-model="dashboardStartTime" :label="$t('dashboard.message05')" clearable /></div>
            <div class="flex flex-col md6"><va-date-input class="mb-2 pr-2" v-model="dashboardStopDate" :month-names="$tm('calendar.monthnames')" :weekday-names="$tm('calendar.weekdaynames')" :first-weekday="$t('calendar.firstweekdays')" :label="$t('dashboard.message04')" clearable /></div>
            <div class="flex flex-col md6"><va-time-input class="mb-2 pl-2" v-model="dashboardStopTime" :label="$t('dashboard.message06')" clearable /></div>
            <div class="flex flex-col md12">
                <va-divider dashed>
                    <span class="px-2">{{ $t('dashboard.message37') }}</span>
                </va-divider>
            </div>
            <div class="flex flex-col md12">
                <va-button-group grow>
                    <va-button v-on:click="buttonClickInterval(3600)" class="mr-1">{{ $t('dashboard.message38') }}</va-button>
                    <va-button v-on:click="buttonClickInterval(10800)" class="mr-1">{{ $t('dashboard.message39') }}</va-button>
                    <va-button v-on:click="buttonClickInterval(21600)" class="mr-1">{{ $t('dashboard.message40') }}</va-button>
                    <va-button v-on:click="buttonClickInterval(43200)" class="mr-1">{{ $t('dashboard.message41') }}</va-button>
                    <va-button v-on:click="buttonClickInterval(86400)" class="mr-1">{{ $t('dashboard.message42') }}</va-button>
                    <va-button v-on:click="buttonClickInterval(259200)">{{ $t('dashboard.message43') }}</va-button>
                </va-button-group>
            </div>
            <div class="flex flex-col md12"><br></div>
            <div class="flex flex-col md12"><va-input class="mb-2" v-model="dashboardInterval" :placeholder="$t('dashboard.message44')" :label="$t('dashboard.message45')" clearable /></div>
        </div>
    </va-modal>
    
    <va-inner-loading :loading="innerLoading">
        <div class="row layout va-gutter-3">
            <div class="flex layout va-gutter-3 md6" v-for="(dashboard, index01) in dashboards" :key="index01">
                <va-modal v-model="showModalRemove[index01]" :title="$t('dashboard.message08')" v-on:ok="putDashboardRemove(index01)" :okText="$t('message.delete')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
                    <div>{{ $t('dashboard.message10') }} {{ index01 + 1 }} ({{ dashboard.name }})</div>
                </va-modal>

                <va-modal v-model="showModalEdit[index01]" :title="$t('dashboard.message09')" v-on:ok="putDashboardUpdate(index01)" :okText="$t('message.save')" :cancelText="$t('message.cancel')" noOutsideDismiss="true" zIndex="5">
                    <div class="row">
                        <div class="flex flex-col md12"><va-input class="mb-2" v-model="dashboardEditName" :placeholder="$t('dashboard.message02')" :label="$t('dashboard.message02')" /></div>
                        <div class="flex flex-col md6"><va-date-input class="mb-2 pr-2" v-model="dashboardEditStartDate" :month-names="$tm('calendar.monthnames')" :weekday-names="$tm('calendar.weekdaynames')" :first-weekday="$t('calendar.firstweekdays')" :label="$t('dashboard.message03')" clearable /></div>
                        <div class="flex flex-col md6"><va-time-input class="mb-2 pl-2" v-model="dashboardEditStartTime" :label="$t('dashboard.message05')" clearable /></div>
                        <div class="flex flex-col md6"><va-date-input class="mb-2 pr-2" v-model="dashboardEditStopDate" :month-names="$tm('calendar.monthnames')" :weekday-names="$tm('calendar.weekdaynames')" :first-weekday="$t('calendar.firstweekdays')" :label="$t('dashboard.message04')" clearable /></div>
                        <div class="flex flex-col md6"><va-time-input class="mb-2 pl-2" v-model="dashboardEditStopTime" :label="$t('dashboard.message06')" clearable /></div>
                        <div class="flex flex-col md12">
                        <va-divider dashed>
                            <span class="px-2">{{ $t('dashboard.message37') }}</span>
                        </va-divider>
                        </div>
                        <div class="flex flex-col md12">
                            <va-button-group grow>
                                <va-button v-on:click="buttonClickEditInterval(3600)" class="mr-1">{{ $t('dashboard.message38') }}</va-button>
                                <va-button v-on:click="buttonClickEditInterval(10800)" class="mr-1">{{ $t('dashboard.message39') }}</va-button>
                                <va-button v-on:click="buttonClickEditInterval(21600)" class="mr-1">{{ $t('dashboard.message40') }}</va-button>
                                <va-button v-on:click="buttonClickEditInterval(43200)" class="mr-1">{{ $t('dashboard.message41') }}</va-button>
                                <va-button v-on:click="buttonClickEditInterval(86400)" class="mr-1">{{ $t('dashboard.message42') }}</va-button>
                                <va-button v-on:click="buttonClickEditInterval(259200)">{{ $t('dashboard.message43') }}</va-button>
                            </va-button-group>
                        </div>
                        <div class="flex flex-col md12"><br></div>
                        <div class="flex flex-col md12"><va-input class="mb-2" v-model="dashboardEditInterval" :placeholder="$t('dashboard.message44')" :label="$t('dashboard.message45')" clearable /></div>
                    </div>
                    <div class="md12">
                        <va-divider dashed>
                            <span class="px-2">{{ $t('dashboard.message11') }}</span>
                        </va-divider>
                    </div>
                    <div class="md12" align="center">
                        <va-button icon="add" v-bind:disabled="userBlock" v-on:click="buttonClickChartNew(index01)" class="mr-3" />
                    </div>
                    <div class="flex layout md12 va-gutter-3" v-for="(chartData, index02) in dashboardsEdit[index01].chartdata" :key="index02">
                        <va-card stripe :stripe-color="stripeColorChart[index02]">
                            <va-card-title>{{ $t('dashboard.message46') }} {{ index02 + 1 }}</va-card-title>
                            <va-card-content>
                                <div class="flex flex-col md12"><va-select class="mb-2" v-model="spaceValue[index01][index02]" :options="spaceOptions" v-on:update:model-value="spaceSelectUpdateValue(index01, index02)" :label="$t('dashboard.message29')" :placeholder="$t('dashboard.message30')" :no-options-text="$t('message.listempty')" searchable></va-select></div>
                                <div class="flex flex-col md12"><va-select class="mb-2" v-model="metricValue[index01][index02]" :options="metricOptions[index01][index02]" :label="$t('dashboard.message31')" :placeholder="$t('dashboard.message32')" :no-options-text="$t('message.listempty')" searchable></va-select></div>
                                <div class="flex flex-col md12"><va-select class="mb-2" v-model="graphtypeValue[index01][index02]" :options="graphtypeOptions" :placeholder="$t('dashboard.message33')" :label="$t('dashboard.message33')" :no-options-text="$t('message.listempty')"></va-select></div>
                                <div class="flex flex-col md12"><va-color-input class="mb-2" v-model="dashboardsEdit[index01].chartdata[index02].graphcolor" :placeholder="$t('dashboard.message34')" :label="$t('dashboard.message34')" /></div>
                                <div class="flex flex-col md12"><va-select class="mb-2" v-model="groupfuncValue[index01][index02]" :options="groupfuncOptions" :placeholder="$t('dashboard.message35')" :label="$t('dashboard.message35')" :no-options-text="$t('message.listempty')"></va-select></div>
                                <div class="flex flex-col md12"><va-input class="mb-2" v-model="dashboardsEdit[index01].chartdata[index02].graphorder" :placeholder="$t('dashboard.message36')" :label="$t('dashboard.message36')" /></div>
                                <div class="md12" align="center">
                                    <va-button icon="delete" v-bind:disabled="userBlock" v-on:click="buttonClickChartRemove(index01, index02)" class="mr-3" />
                                </div>
                            </va-card-content>
                        </va-card>
                    </div>
                    <div class="md12">
                        <va-divider dashed />
                    </div>
                </va-modal>

                <va-card :color="cardColor">
                    <va-card-title>
                        <div class="layout va-gutter-1 md6 align-content-center justify-start">{{ index01 + 1 }} {{ dashboardTitle[index01] }}</div>
                        <div class="row layout va-gutter-1 va-spacing-x-1 md6 justify-end">
                            <va-button icon="edit" round v-on:click="putDashboardUpdateForm(index01)" />
                            <va-button icon="delete" color="danger" round v-on:click="putDashboardRemoveConfirm(index01)" />
                        </div>
                    </va-card-title>
                    <va-card-content>
                            <Bar v-if="barLoading" :options="chartOptions" :data="chartDataArr[index01]" chart-id="index" height="300vh"></Bar>
                    </va-card-content>
                </va-card>
            </div>
        </div>
    </va-inner-loading>
</template>

<script>
    import $ from 'jquery'
    import moment from 'moment'
    import { mapState } from 'vuex'
    import { Bar } from 'vue-chartjs'
    import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, LineElement, CategoryScale, LinearScale, LineController, PointElement } from 'chart.js'

    ChartJS.register(Title, Tooltip, Legend, BarElement, LineElement, CategoryScale, LinearScale, LineController, PointElement)

    export default {
        components: { Bar },
        chartDataIntervalID: null,
        data() {
            const graphtypeOptions = [
                {text: "Bar", value: "bar"},
                {text: "Line", value: "line"},
            ];
            const groupfuncOptions = [
                {text: "Avg", value: "avg"},
            ];

            return {
                cardColorL: "#207ba5ff",
                cardColorD: "#1f303eff",
                cardColor: "",

                spaceOptions: [],
                spaceValue: [],
                metricOptions: [],
                metricValue: [],
                graphtypeOptions: graphtypeOptions,
                graphtypeValue: [],
                groupfuncOptions: groupfuncOptions,
                groupfuncValue: [],

                dashboardTitle: [],
                dashboardName: "",
                dashboardStartDate: undefined,
                dashboardStartTime: undefined,
                dashboardStopDate: undefined,
                dashboardStopTime: undefined,
                dashboardInterval: undefined,

                dashboardEditName: "",
                dashboardEditStartDate: undefined,
                dashboardEditStartTime: undefined,
                dashboardEditStopDate: undefined,
                dashboardEditStopTime: undefined,
                dashboardEditInterval: undefined,

                /* Переменные состояния модальных окон */
                showModalCreate: false,
                showModalRemove: [],
                showModalEdit: [],

                stripeColorChart: [],

                dashboards: [],
                dashboardsEdit: [],
                innerLoading:true,
                barLoading: false,
                chartDataArr: [],
                chartOptions: {
                    responsive: true,
                    maintainAspectRatio: false,
                    color: "#ffffff",
                    borderColor: "#ffffff",
                    scales: {
                        x: {
                            border: {
                                width: 0,
                            },
                            grid: {
                                color: "#ffffff88",
                                display: true,
                            },
                            ticks: {
                                color: "#ffffff",
                            },
                        },
                        y: {
                            border: {
                                width: 0,
                            },
                            grid: {
                                color: "#ffffff88",
                                display: true,
                            },
                            ticks: {
                                color: "#ffffff",
                            },
                        },
                    }
                },
            }
        },
        methods: {
            buttonClickInterval(sec) {
                this.dashboardInterval = sec;
                this.dashboardStartDate = undefined;
                this.dashboardStartTime = undefined;
            },
            buttonClickEditInterval(sec) {
                this.dashboardEditInterval = sec;
                this.dashboardEditStartDate = undefined;
                this.dashboardEditStartTime = undefined;
            },
            buttonClickNew() {
                this.dashboardName = "";
                this.dashboardStartDate = undefined;
                this.dashboardStopDate = undefined;
                this.dashboardStartTime = undefined;
                this.dashboardStopTime = undefined;
                this.dashboardInterval = undefined;

                this.showModalCreate = true;
            },
            buttonClickChartNew(index) {
                let chartSize = 0;
                if (this.dashboardsEdit[index].chartdata == null) {
                    this.dashboardsEdit[index].chartdata = [];
                    chartSize = this.dashboardsEdit[index].chartdata.push(
                        {
                            id: "",
                            metric: "",
                            spaceid: "",
                            graphtype: "",
                            graphcolor: "#ade2ffbe",
                            groupfunc: "",
                            graphorder: 0
                        }
                    );

                    let chart01 = [];
                    chart01[0] = [];
                    this.metricOptions[index] = chart01;
                    let chart02 = [];
                    chart02[0] = "";
                    this.metricValue[index] = chart02;
                    let chart03 = [];
                    chart03[0] = "";
                    this.spaceValue[index] = chart03;
                    let chart04 = [];
                    chart04[0] = "";
                    this.graphtypeValue[index] = chart04;
                    let chart05 = [];
                    chart05[0] = "";
                    this.groupfuncValue[index] = chart05;
                } else {
                    chartSize = this.dashboardsEdit[index].chartdata.push(
                        {
                            id: "",
                            metric: "",
                            spaceid: "",
                            graphtype: "",
                            graphcolor: "#ade2ffbe",
                            groupfunc: "",
                            graphorder: 0
                        }
                    );

                    this.metricOptions[index][chartSize - 1] = [];
                    this.metricValue[index][chartSize -1] = "";
                    this.spaceValue[index][chartSize -1] = "";
                    this.graphtypeValue[index][chartSize -1] = "";
                    this.groupfuncValue[index][chartSize -1] = "";
                }

                this.stripeColorChart[chartSize - 1] = "warning";
            },
            buttonClickChartRemove(index01, index02) {
                let vm = this;
                vm.dashboardsEdit[index01].chartdata.splice(index02, 1);

                if (this.dashboardsEdit[index01].chartdata != null) {
                    this.dashboardsEdit[index01].chartdata.forEach(function(item, i) {
                        // Переопределение выбора значений для полей SELECT после сдвига массива
                        vm.spaceOptions.forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index01].chartdata[i].spaceid) {
                                vm.spaceValue[index01][i] = item;
                                vm.spaceSelectUpdateValue(index01, i);
                                return
                            }
                        });
                        vm.metricOptions[index01][i].forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index01].chartdata[i].metric) {
                                vm.metricValue[index01][i] = item;
                                return
                            }
                        });
                        vm.graphtypeOptions.forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index01].chartdata[i].graphtype) {
                                vm.graphtypeValue[index01][i] = item;
                                return
                            }
                        });
                        vm.groupfuncOptions.forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index01].chartdata[i].groupfunc) {
                                vm.groupfuncValue[index01][i] = item;
                                return
                            }
                        });
                    });
                }
            },
            putDashboardRemoveConfirm(index) {
                this.showModalRemove[index] = true;
            },
            putDashboardUpdateForm(index) {
                let vm = this;

                this.dashboardsEdit = JSON.parse(JSON.stringify(this.dashboards));

                this.dashboardEditName = this.dashboardsEdit[index].name;

                if (this.dashboardsEdit[index].starttime == 0 || this.dashboardsEdit[index].starttime == undefined) {
                    this.dashboardEditStartDate = undefined;
                    this.dashboardEditStartTime = undefined;
                } else {
                    this.dashboardEditStartDate = new Date(this.dashboardsEdit[index].starttime);
                    this.dashboardEditStartTime = new Date(this.dashboardsEdit[index].starttime);
                }
                if (this.dashboardsEdit[index].stoptime == 0 || this.dashboardsEdit[index].stoptime == undefined) {
                    this.dashboardEditStopDate = undefined;
                    this.dashboardEditStopTime = undefined;
                } else {
                    this.dashboardEditStopDate = new Date(this.dashboardsEdit[index].stoptime);
                    this.dashboardEditStopTime = new Date(this.dashboardsEdit[index].stoptime);
                }

                if (this.dashboardsEdit[index].interval == 0) {
                    this.dashboardEditInterval = undefined;
                } else {
                    this.dashboardEditInterval = this.dashboardsEdit[index].interval;
                }

                /* Инициализация metricOptions и metricValue для формы */
                if (vm.dashboardsEdit[index].chartdata != null) {
                    let chart01 = [];
                    let chart02 = [];
                    let chart03 = [];
                    let chart04 = [];
                    let chart05 = [];
                    vm.dashboardsEdit[index].chartdata.forEach(function(item, i) {
                        chart01[i] = [];
                        chart02[i] = "";
                        chart03[i] = "";
                        chart04[i] = "";
                        chart05[i] = "";
                    });
                    vm.metricOptions[index] = chart01;
                    vm.metricValue[index] = chart02;
                    vm.spaceValue[index] = chart03;
                    vm.graphtypeValue[index] = chart04;
                    vm.groupfuncValue[index] = chart05;
                }

                if (this.dashboards[index].chartdata != null) {
                    this.dashboards[index].chartdata.forEach(function(item, i) {
                        vm.stripeColorChart[i] = "success";

                        vm.spaceOptions.forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index].chartdata[i].spaceid) {
                                vm.spaceValue[index][i] = item;
                                return
                            }
                        });
                        vm.metricOptions[index][i].forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index].chartdata[i].metric) {
                                vm.metricValue[index][i] = item;
                                return
                            }
                        });
                        vm.graphtypeOptions.forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index].chartdata[i].graphtype) {
                                vm.graphtypeValue[index][i] = item;
                                return
                            }
                        });
                        vm.groupfuncOptions.forEach(function(item) {
                            if (item.value == vm.dashboardsEdit[index].chartdata[i].groupfunc) {
                                vm.groupfuncValue[index][i] = item;
                                return
                            }
                        });

                        if (vm.dashboardsEdit[index].chartdata[i].spaceid.value != "") {
                            vm.spaceSelectUpdateValue(index, i)
                        }

                        if (vm.dashboardsEdit[index].chartdata[i].graphcolor == "") {
                            vm.dashboardsEdit[index].chartdata[i].graphcolor = "#ade2ffbe";
                        }
                    });
                }

                this.showModalEdit[index] = true;
            },
            putDashboardUpdate(index) {
                let vm = this;
                let dashboardStartDateTime = parseInt(0);
                let dashboardStopDateTime = parseInt(0);
                let dashboardIntervalTime = parseInt(0);

                if (vm.dashboardEditStartDate != undefined && vm.dashboardEditStartTime == undefined) {
                    vm.showModalEdit[index] = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message12'), color: 'warning' });
                    return;
                }
                if (vm.dashboardEditStartDate == undefined && vm.dashboardEditStartTime != undefined) {
                    vm.showModalEdit[index] = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message13'), color: 'warning' });
                    return;
                }
                if (vm.dashboardEditStopDate != undefined && vm.dashboardEditStopTime == undefined) {
                    vm.showModalEdit[index] = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message14'), color: 'warning' });
                    return;
                }
                if (vm.dashboardEditStopDate == undefined && vm.dashboardEditStopTime != undefined) {
                    vm.showModalEdit[index] = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message15'), color: 'warning' });
                    return;
                }

                if (vm.dashboardEditStartDate != undefined && vm.dashboardEditStartTime != undefined) {
                    let dashboardStartJSDateTime = new Date(vm.dashboardEditStartDate.getFullYear(), vm.dashboardEditStartDate.getMonth(), vm.dashboardEditStartDate.getDate(), vm.dashboardEditStartTime.getHours(), vm.dashboardEditStartTime.getMinutes());
                    dashboardStartDateTime = parseInt((dashboardStartJSDateTime.getTime()));
                }
                if (vm.dashboardEditStopDate != undefined && vm.dashboardEditStopTime != undefined) {
                    let dashboardStopJSDateTime = new Date(vm.dashboardEditStopDate.getFullYear(), vm.dashboardEditStopDate.getMonth(), vm.dashboardEditStopDate.getDate(), vm.dashboardEditStopTime.getHours(), vm.dashboardEditStopTime.getMinutes());
                    dashboardStopDateTime = parseInt((dashboardStopJSDateTime.getTime()));
                }

                if (dashboardStartDateTime > 0 && dashboardStopDateTime > 0) {
                    if (dashboardStopDateTime - dashboardStartDateTime < 0) {
                        vm.showModalEdit[index] = true;
                        vm.$vaToast.init({ message: vm.$t('dashboard.message16'), color: 'warning' });
                        return;
                    }
                    if (dashboardStopDateTime - dashboardStartDateTime > 3 * 31 * 24 * 60 * 60 * 1000) {
                        vm.showModalEdit[index] = true;
                        vm.$vaToast.init({ message: vm.$t('dashboard.message17'), color: 'warning' });
                        return;
                    }
                }
                if (vm.dashboardEditInterval == undefined || vm.dashboardEditInterval == 0 || vm.dashboardEditInterval == "" || isNaN(parseInt(vm.dashboardEditInterval))) {
                    dashboardStartDateTime = parseInt(0);
                    dashboardIntervalTime = parseInt(0)
                } else {
                    if (vm.dashboardEditInterval < 0 || /^\d+$/.test(vm.dashboardEditInterval) == false) {
                        vm.showModalEdit[index] = true;
                        vm.$vaToast.init({ message: vm.$t('dashboard.message47'), color: 'warning' });
                        return;
                    } else {
                        dashboardIntervalTime = parseInt(vm.dashboardEditInterval);
                    }
                }

                let errorFlag = 0;
                if (vm.dashboardsEdit[index].chartdata != undefined) {
                    vm.dashboardsEdit[index].chartdata.forEach(function(item, i) {
                        vm.dashboardsEdit[index].chartdata[i].graphorder = parseInt(vm.dashboardsEdit[index].chartdata[i].graphorder);
                        vm.dashboardsEdit[index].chartdata[i].spaceid = vm.spaceValue[index][i].value;
                        vm.dashboardsEdit[index].chartdata[i].metric = vm.metricValue[index][i].value;
                        vm.dashboardsEdit[index].chartdata[i].graphtype = vm.graphtypeValue[index][i].value;
                        vm.dashboardsEdit[index].chartdata[i].groupfunc = vm.groupfuncValue[index][i].value;
                    
                        if (vm.dashboardsEdit[index].chartdata[i].spaceid == undefined) {
                            vm.showModalEdit[index] = true;
                            vm.$vaToast.init({ message: vm.$t('dashboard.message18') + ' ' + (i + 1), color: 'warning' });
                            errorFlag = 1;
                            return;
                        }
                        if (vm.dashboardsEdit[index].chartdata[i].metric == undefined) {
                            vm.showModalEdit[index] = true;
                            vm.$vaToast.init({ message: vm.$t('dashboard.message19') + ' ' + (i + 1), color: 'warning' });
                            errorFlag = 1;
                            return;
                        }
                        if (vm.dashboardsEdit[index].chartdata[i].graphtype == undefined) {
                            vm.showModalEdit[index] = true;
                            vm.$vaToast.init({ message: vm.$t('dashboard.message20') + ' ' + (i + 1), color: 'warning' });
                            errorFlag = 1;
                            return;
                        }
                        if (vm.dashboardsEdit[index].chartdata[i].groupfunc == undefined) {
                            vm.showModalEdit[index] = true;
                            vm.$vaToast.init({ message: vm.$t('dashboard.message21') + ' ' + (i + 1), color: 'warning' });
                            errorFlag = 1;
                            return;
                        }
                    });
                }

                if (errorFlag == 1) {
                    return;
                }

                let dataPut = {
                    id: vm.dashboardsEdit[index].id,
                    name: vm.dashboardEditName,
                    starttime: parseInt(dashboardStartDateTime),
                    stoptime: parseInt(dashboardStopDateTime),
                    interval: parseInt(dashboardIntervalTime),
                    status: 1,
                    chartdata: vm.dashboardsEdit[index].chartdata
                };

                $.ajax({
                    url: "/api/v1/admin/dashboard/edit",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('dashboard.message22'), color: 'primary' });
                            vm.showModalEdit[index] = false;
                            vm.innerLoading = true;
                            vm.barLoading = false;
                            vm.getDashboard();
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('dashboard.message23'), color: 'danger' });
                            vm.showModalEdit[index] = true;
                            return true;
                        },
                    }
                });
            },
            putDashboardRemove(index) {
                let vm = this;

                $.ajax({
                    url: "/api/v1/admin/dashboard/remove",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(vm.dashboards[index]),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('dashboard.message24'), color: 'primary' });
                            vm.showModalRemove[index] = false;
                            vm.innerLoading = true;
                            vm.barLoading = false;
                            vm.getDashboard();
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('dashboard.message25'), color: 'danger' });
                            vm.showModalRemove[index] = true;
                            return true;
                        },
                    }
                });
            },
            putDashboardCreate() {
                let vm = this;
                let dashboardStartDateTime = parseInt(0);
                let dashboardStopDateTime = parseInt(0);
                let dashboardIntervalTime = parseInt(0);

                if (vm.dashboardName == "") {
                    vm.showModalCreate = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message26'), color: 'danger' });
                    return;
                }
                if (vm.dashboardStartDate != undefined && vm.dashboardStartTime == undefined) {
                    vm.showModalCreate = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message12'), color: 'danger' });
                    return;
                }
                if (vm.dashboardStartDate == undefined && vm.dashboardStartTime != undefined) {
                    vm.showModalCreate = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message13'), color: 'danger' });
                    return;
                }
                if (vm.dashboardStopDate != undefined && vm.dashboardStopTime == undefined) {
                    vm.showModalCreate = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message14'), color: 'danger' });
                    return;
                }
                if (vm.dashboardStopDate == undefined && vm.dashboardStopTime != undefined) {
                    vm.showModalCreate = true;
                    vm.$vaToast.init({ message: vm.$t('dashboard.message15'), color: 'danger' });
                    return;
                }
                
                if (vm.dashboardStartDate != undefined && vm.dashboardStartTime != undefined) {
                    let dashboardStartJSDateTime = new Date(vm.dashboardStartDate.getFullYear(), vm.dashboardStartDate.getMonth(), vm.dashboardStartDate.getDate(), vm.dashboardStartTime.getHours(), vm.dashboardStartTime.getMinutes());
                    dashboardStartDateTime = parseInt((dashboardStartJSDateTime.getTime()));
                }
                if (vm.dashboardStopDate != undefined && vm.dashboardStopTime != undefined) {
                    let dashboardStopJSDateTime = new Date(vm.dashboardStopDate.getFullYear(), vm.dashboardStopDate.getMonth(), vm.dashboardStopDate.getDate(), vm.dashboardStopTime.getHours(), vm.dashboardStopTime.getMinutes());
                    dashboardStopDateTime = parseInt((dashboardStopJSDateTime.getTime()));
                }

                if (dashboardStartDateTime > 0 && dashboardStopDateTime > 0) {
                    if (dashboardStopDateTime - dashboardStartDateTime < 0) {
                        vm.showModalCreate = true;
                        vm.$vaToast.init({ message: vm.$t('dashboard.message16'), color: 'warning' });
                        return;
                    }
                    if (dashboardStopDateTime - dashboardStartDateTime > 3 * 31 * 24 * 60 * 60 * 1000) {
                        vm.showModalCreate = true;
                        vm.$vaToast.init({ message: vm.$t('dashboard.message47'), color: 'warning' });
                        return;
                    }
                }

                if (vm.dashboardInterval == undefined || vm.dashboardInterval == 0 || vm.dashboardInterval == "" || isNaN(parseInt(vm.dashboardInterval))) {
                    dashboardStartDateTime = 0;
                    dashboardIntervalTime = parseInt(0)
                } else {
                    if (vm.dashboardInterval < 0 || /^\d+$/.test(vm.dashboardInterval) == false) {
                        vm.showModalCreate = true;
                        vm.$vaToast.init({ message: vm.$t('dashboard.message47'), color: 'warning' });
                        return;
                    } else {
                        dashboardIntervalTime = parseInt(vm.dashboardInterval);
                    }
                }

                let dataPut = {
                    name: vm.dashboardName,
                    starttime: parseInt(dashboardStartDateTime),
                    stoptime: parseInt(dashboardStopDateTime),
                    interval: parseInt(dashboardIntervalTime),
                    status: 1,
                };
                $.ajax({
                    url: "/api/v1/admin/dashboard/create",
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    statusCode: {
                        200: function() {
                            vm.$vaToast.init({ message: vm.$t('dashboard.message27'), color: 'primary' });
                            vm.showModalCreate = false;
                            vm.innerLoading = true;
                            vm.barLoading = false;
                            vm.getDashboard();
                            return true;
                        },
                        500: function() {
                            vm.$vaToast.init({ message: vm.$t('dashboard.message28'), color: 'danger' });
                            vm.showModalCreate = true;
                            return true;
                        },
                    }
                });
            },
            getDashboard() {
                let vm = this;
                $.ajax({
                    url: "/api/v1/admin/dashboard/select?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        if (data == null) {
                            data = [];
                            vm.dashboards = [];
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
                        vm.dashboards = data;
                        data.forEach(function(item, i) {
                            vm.showModalRemove[i] = false;
                            vm.showModalEdit[i] = false;

                            let dateStartTitle = "-30m";
                            let dateStopTitle = "Now";
                            if (vm.dashboards[i].starttime > 0) {
                                dateStartTitle = moment(vm.dashboards[i].starttime).format("DD.MM.YYYY HH:mm");
                            }
                            if (vm.dashboards[i].stoptime > 0) {
                                dateStopTitle = moment(vm.dashboards[i].stoptime).format("DD.MM.YYYY HH:mm");
                            }

                            /* Формирование заголовка дашборда */
                            let dateTitle = "";
                            if (vm.dashboards[i].stoptime == 0) {
                                dateTitle = " (Now)";
                            } else if (vm.dashboards[i].starttime > 0) {
                                dateTitle = " (" + dateStartTitle + " - " + dateStopTitle + ")";
                            }
                            vm.dashboardTitle[i] = vm.dashboards[i].name + dateTitle;
                        });
                        vm.putChartData();
                    },
                });
            },
            putChartData() {
                let vm = this;
                if (vm.dashboards == null || vm.dashboards.length == 0) {
                    vm.innerLoading = false;
                    return;
                }
                $.ajax({
                    url: "/api/v1/admin/dashboard/data?" + Math.random(),
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(vm.dashboards),
                    success: function (data) {
                        data.forEach(function(item01, i01) {
                            if (item01.labels != null) {
                                item01.labels.forEach(function(item02, i02) {
                                    if (item01.stoptime - item01.starttime <= (180 * 1000)) {
                                        data[i01].labels[i02] = moment(item02).format("HH:mm:ss");
                                    } else {
                                        data[i01].labels[i02] = moment(item02).format("DD.MM.YYYY HH:mm");
                                    }
                                });
                            }
                            if (item01.datasets == null) {
                                data[i01].datasets = [];
                            }
                        });
                        vm.chartDataArr = data;
                        vm.innerLoading = false;
                        vm.barLoading = true;
                    },
                });
            },
            chartDataUpdate() {
                this.putChartData();
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
            spaceSelectUpdateValue(index01, index02) {
                let vm = this;
                let dataPut = {
                    spaceid: this.spaceValue[index01][index02].value,
                };
                if (this.spaceValue[index01][index02].value != "") {
                    $.ajax({
                        url: "/api/v1/admin/rawdata/metric/select?" + Math.random(),
                        type: "PUT",
                        dataType: "json",
                        data: JSON.stringify(dataPut),
                        success: function (data) {
                            if (data == null) {
                                data = [];
                                vm.metricOptions[index01][index02] = [];
                                vm.metricValue[index01][index02] = "";
                            } else {
                                if (data.length > 0) {
                                    data.forEach(function(item, i) {
                                        vm.metricOptions[index01][index02][i] = {text: item.metric, value: item.metric};
                                    });

                                    vm.metricOptions[index01][index02].forEach(function(item) {
                                        if (item.value == vm.dashboardsEdit[index01].chartdata[index02].metric) {
                                            vm.metricValue[index01][index02] = item;
                                            return
                                        }
                                    });
                                } else {
                                    vm.metricOptions[index01][index02] = [];
                                    vm.metricValue[index01][index02] = "";
                                }
                            }
                            return true;
                        }
                    });
                }
            },
        },  
        created() {
            if (this.$store.state.mode == false) {
                this.cardColor = this.cardColorL;
            } else {
                this.cardColor = this.cardColorD;
            }

            this.getDashboard();
            this.spaceSelect();
            this.chartDataIntervalID = window.setInterval(this.chartDataUpdate, 10000);
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
                    this.cardColor = this.cardColorL;

                } else {
                    this.cardColor = this.cardColorD;
                }
            }
        }
    }
</script>
<style>
    * {
        --va-card-outlined-border: 1px solid var(--va-background-element);
    }
</style>