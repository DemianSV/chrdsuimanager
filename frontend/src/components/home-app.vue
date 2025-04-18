<template>
<div class="row layout va-gutter-3">
    <div class="flex layout va-gutter-3 md3">
        <va-card :color="cardColor02">
            <va-card-title> {{ $t('message.time') }} </va-card-title>
            <va-card-content><h5 class="va-h5"> {{ currentDT }} </h5></va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md3">
        <va-card to="/user" v-bind:disabled="userBlock" :color="cardColor01">
            <va-card-title> {{ $t('message.users') }} </va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="manage_accounts" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md3">
        <va-card to="/space" :color="cardColor01">
            <va-card-title> {{ $t('message.spaces') }} </va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="workspaces" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md3">
        <va-card to="/registration" :color="cardColor01">
            <va-card-title> {{ $t('message.modules') }} </va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="dns" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md2">
        <va-card to="/task" :color="cardColor01">
            <va-card-title> {{ $t('message.tasks') }} </va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="add_task" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md2">
        <va-card to="/problem" v-bind:disabled="userBlock" :color="cardColor01">
            <va-card-title> {{ $t('message.problems') }} </va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="report_problem" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md2">
        <va-card to="/dashboard" :color="cardColor01">
            <va-card-title> {{ $t('message.dashboards') }} </va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="dashboard_customize" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md2">
        <va-card v-bind:color="databaseColor" to="/database" v-bind:disabled="adminBlock">
            <va-card-title> {{ $t('message.statusdb') }} </va-card-title>
            <va-card-content><h5 class="va-h5"> {{ databaseData }} </h5></va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md2">
        <va-card v-bind:color="datamanagerColor">
            <va-card-title> {{ $t('message.statusdm') }} </va-card-title>
            <va-card-content><h5 class="va-h5"> {{ datamanagerData }} </h5></va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md2">
        <va-card v-bind:color="uimanagerColor">
            <va-card-title> {{ $t('message.statusuim') }} </va-card-title>
            <va-card-content><h5 class="va-h5"> {{ uimanagerData }} </h5></va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md6">
        <va-card to="/rawdata" :color="cardColor01">
            <va-card-title> {{ $t('message.metricsld') }} </va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="table_chart" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md6">
        <va-card to="/rawtext" :color="cardColor01">
            <va-card-title> {{ $t('message.logsld') }}</va-card-title>
            <va-card-content>
                <va-icon class="mr-2" name="table_chart" size="3rem" />
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md6">
        <va-card to="/rawdata" :color="cardColor02">
            <va-card-title> {{ $t('message.metricsr') }} </va-card-title>
            <va-card-content>
                <Bar :options="chartOptions" :data="chartData01" chart-id="barChart01"></Bar>
            </va-card-content>
        </va-card>
    </div>
    <div class="flex layout va-gutter-3 md6">
        <va-card to="/rawtext" :color="cardColor02">
            <va-card-title> {{ $t('message.logsr') }} </va-card-title>
            <va-card-content>
                <Bar :options="chartOptions" :data="chartData02" chart-id="barChart02"></Bar>
            </va-card-content>
        </va-card>
    </div>
</div>
</template>

<script>
    import $ from 'jquery'
    import moment from 'moment'
    import { Bar } from 'vue-chartjs'
    import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale } from 'chart.js'
    ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

    export default {
        name: 'main-app',
        components: { Bar },
        uimanagerIntervalID: null,
        datamanagerIntervalID: null,
        databaseIntervalID: null,
        createDTID: null,
        chartDataIntervalID: null,
        data() {
            return {
                cardColorL01: "#ade2ffff",
                cardColorL02: "#207ba5ff",
                cardColorD01: "#36536bff",
                cardColorD02: "#1f303eff",
                cardColor01: "",
                cardColor02: "",
                currentDT: this.$t('message.nodata'),
                tab: "Home",
                userRole: null,
                userBlock: true,
                adminBlock: true,
                databaseData: this.$t('message.nodata'),
                databaseColor: "",
                uimanagerData: this.$t('message.nodata'),
                uimanagerColor: "",
                datamanagerData: this.$t('message.nodata'),
                datamanagerColor: "",
                chartData01: {
                    labels: [],
                    datasets: [
                        {
                            label: '',
                            backgroundColor: '',
                            data: [],
                            barPercentage: 1
                        }
                    ]
                },
                chartData02: {
                    labels: [],
                    datasets: [
                        {
                            label: '',
                            backgroundColor: '',
                            data: [],
                            barPercentage: 1,
                        }
                    ]
                },
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
                    },
                    plugins: {
                        legend: {
                            display: false
                        }
                    },
                },
            }
        },
        methods: {
            chartDataAPI(dataSRC) {
                const vm = this;
                let dataPut = {
                    datasrc: dataSRC,
                };
                $.ajax({
                    url: "/api/v1/admin/chartdata?" + Math.random(),
                    type: "PUT",
                    dataType: "json",
                    data: JSON.stringify(dataPut),
                    success: function (data) {
                        if (data.labels != null) {
                            data.labels.forEach(function(item, i) {
                                data.labels[i] = moment(item).format("HH:mm");
                            });
                            if (dataSRC == "raw_data") {
                                vm.chartData01 = data;
                            }
                            if (dataSRC == "raw_text") {
                                vm.chartData02 = data;
                            }
                        }
                        return true;
                    },
                });
            },
            async init() {
                const vm = this;

                if (this.$store.state.mode == false) {
                    this.cardColor01 = this.cardColorL01;
                    this.cardColor02 = this.cardColorL02;
                    this.databaseColor = this.cardColor02;
                    this.uimanagerColor = this.cardColor02;
                    this.datamanagerColor = this.cardColor02;
                } else {
                    this.cardColor01 = this.cardColorD01;
                    this.cardColor02 = this.cardColorD02;
                    this.databaseColor = this.cardColor02;
                    this.uimanagerColor = this.cardColor02;
                    this.datamanagerColor = this.cardColor02;
                }

                try {
                    await $.ajax({
                        url: "/api/v1/userinfo?" + Math.random(),
                        type: "GET",
                        dataType: "json",
                        success: function (data) {
                            vm.userRole = data.role;
                            if (data.role == "superadmin") {
                                vm.userBlock = false
                                vm.adminBlock = false
                            } else if (data.role == "admin") {
                                vm.userBlock = false
                                vm.adminBlock = true
                                vm.databaseData = vm.$t('message.noaccess')
                            }
                            else {
                                vm.userBlock = true
                                vm.adminBlock = true
                                vm.databaseData = vm.$t('message.noaccess')
                                vm.userData = vm.$t('message.noaccess')
                            }
                            return true;
                        },
                    });
                } catch(err) {
                    console.log("UIManager Error!");
                }

                console.log("Role: " + this.userRole);
                this.chartDataAPI("raw_data");
                this.chartDataAPI("raw_text");
                this.databaseStatus();
                this.uimanagerStatus();
                this.datamanagerStatus();
                this.createDTID = window.setInterval(this.createDT, 1000);
                this.uimanagerIntervalID = window.setInterval(this.uimanagerStatus, 10000);
                this.datamanagerIntervalID = window.setInterval(this.datamanagerStatus, 10000);
                this.databaseIntervalID = window.setInterval(this.databaseStatus, 10000);
                this.chartDataIntervalID = window.setInterval(this.chartDataUpdate, 10000);
            },
            chartDataUpdate() {
                this.chartDataAPI("raw_data");
                this.chartDataAPI("raw_text");
            },
            databaseStatus() {
                const vm = this;
                $.ajax({
                    url: "/api/v1/admin/database/status?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        if (data != null) {
                            let status01 = 0;
                            let status02 = 0;
                            data.forEach(function(item) {
                                if (item.up) {
                                    status01++;
                                } else {
                                    status02++;
                                }
                            });
                            vm.databaseData = status01 + " " + vm.$t('message.of') + " " + (status01 + status02);
                            if (status02 > 0) {
                                vm.databaseColor = "danger";
                            } else {
                                vm.databaseColor = "success";
                            }
                        }
                        return true;
                    }
                });
            },
            uimanagerStatus() {
                const vm = this;
                $.ajax({
                    url: "/api/v1/version?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    statusCode: {
                        200: function() {
                            vm.uimanagerColor = "success";
                            vm.uimanagerData = vm.$t('message.available');
                            return true;
                        },
                        401: function() {
                            vm.uimanagerColor = "warning";
                            vm.uimanagerData = vm.$t('message.noaccess');
                            vm.databaseColor = "warning"
                            vm.datamanagerColor = "warning";
                            return true;
                        },
                    },
                    error: function() {
                        vm.uimanagerColor = "danger";
                        vm.uimanagerData = vm.$t('message.notavailable');
                        vm.databaseColor = "warning"
                        vm.datamanagerColor = "warning";
                        return false;
                    }
                });
            },
            datamanagerStatus() {
                const vm = this;
                $.ajax({
                    url: "/api/v1/admin/datamanager/status?" + Math.random(),
                    type: "GET",
                    dataType: "json",
                    success: function (data) {
                        if (data != null) {
                            let status01 = 0;
                            let status02 = 0;
                            data.forEach(function(item) {
                                if (item.up) {
                                    status01++;
                                } else {
                                    status02++;
                                }
                            });
                            vm.datamanagerData = status01 + " " + vm.$t('message.of') + " " + (status01 + status02);
                            if (status02 > 0 && status01 > 0) {
                                vm.datamanagerColor = "warning";
                            } else if (status02 > 0 && status01 == 0) {
                                vm.datamanagerColor = "danger";
                            } else if (status02 == 0 && status01 == 0) {
                                vm.datamanagerColor = "danger";
                            } else {
                                vm.datamanagerColor = "success";
                            }
                        }
                        return true;
                    }
                });
            },
            createDT() {
                this.currentDT = moment().format('DD.MM.YYYY hh:mm:ss');
            }
        },
        created() {
            this.init();
        },
        unmounted() {
            clearInterval(this.uimanagerIntervalID);
            clearInterval(this.datamanagerIntervalID);
            clearInterval(this.databaseIntervalID);
            clearInterval(this.createDTID);
            clearInterval(this.chartDataIntervalID);
        },
        computed: {
            getMode() {
                return this.$store.state.mode;
            }
        },
        watch: {
            getMode(newMode) {
                if (newMode == false) {
                    this.cardColor01 = this.cardColorL01;
                    this.cardColor02 = this.cardColorL02;

                } else {
                    this.cardColor01 = this.cardColorD01;
                    this.cardColor02 = this.cardColorD02;
                }
            }
        }
    }
</script>

<style>
    #barChart01, #barChart02 {
        max-height: 200px;
    }
</style>