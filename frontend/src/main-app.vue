<template>
    <div id="preloader" v-if="showPreloader">
        <div>
            {{ $t('message.loading') }} ...
        </div>
    </div>

    <div class="main" style="height: 100%; width: 100%">
        <va-navbar :color="navBarColor" class="up" id="navbar" shadowed>
            <template #left>
                <va-navbar-item class="logo" v-on:click="this.$router.push('/')"><img :src="logo02" height="50" width="50"></va-navbar-item>
                <va-navbar-item class="version" v-on:click="this.$router.push('/')"> {{ version }} </va-navbar-item>
                <va-navbar-item class="badge" v-on:click="this.$router.push('/')"><va-badge :text="$t('message.release')" color="warning"></va-badge></va-navbar-item>
            </template>
            <template #right>
                <va-navbar-item class="menu">
                    <va-switch v-model="switchValue" v-on:input="switchMode()" color="#0f1a23ff" off-color="#ffd300" style="--va-switch-checker-background-color: #0f1a23ff; --va-switch-checker-active-background-color: #ffffff;">
                        <template #innerLabel>
                            <div class="va-text-center">
                                <va-icon :name="switchValue ? 'dark_mode' : 'light_mode'" :color="switchValue ? '#ffffff' : '#0f1a23ff'" />
                            </div>
                        </template>
                    </va-switch>
                </va-navbar-item>
                <va-navbar-item v-if="!userBlock" class="menu" v-on:click="this.$router.push('/settings')"><va-icon class="mr-2" name="settings" size="2rem" /></va-navbar-item>
                <va-navbar-item class="menu" v-on:click="this.$router.push('/about')"> {{ $t('message.help') }} </va-navbar-item>
                <va-navbar-item class="menu" v-on:click="exit()"> {{ $t('message.exit') }} </va-navbar-item>
                <va-navbar-item class="user"><va-avatar v-bind:color="userRoleColor" v-on:click="showModal=true"> {{ user[0] }} </va-avatar></va-navbar-item>
            </template>
        </va-navbar>
        <div class="down" id="va-app-bar-shadow" style="height: 100%; width: 100%">
            <div class="content" id="content" style="height: 100%; width: 100%"><router-view></router-view></div>
        </div>
    </div>

    <va-modal v-model="showModal" :title="$t('message.settings')" v-on:ok="putPassword()" @before-open="openModalProfile()" :okText="$t('message.save')" :cancelText="$t('message.cancel')" zIndex="5">
        <va-radio v-model="$i18n.locale" :options="localeOptions" :label="$t('message.locale')" :placeholder="$t('message.locale')" :update:modelValue="localeUpdate()" />
        <va-divider />
        <va-input type="password" class="mb-4" v-on:focus="this.isError1=false" v-bind:error="isError1" v-model="curPassword" :placeholder="$t('password.current')" :label="$t('password.current')"></va-input>
        <va-input type="password" class="mb-4" v-on:focus="this.isError2=false" v-bind:error="isError2" v-model="newPassword" :placeholder="$t('password.new')" :label="$t('password.new')"></va-input>
        <va-input type="password" class="mb-4" v-on:focus="this.isError3=false" v-bind:error="isError3" v-model="newPassword2" :placeholder="$t('password.again')" :label="$t('password.new')"></va-input>
    </va-modal>
</template>

<script>
import sha256 from 'sha256'
import $ from 'jquery'
import router from './router.js';
import { useColors } from "vuestic-ui";
import logo02 from '@/assets/logo02.svg';

import { themeQuartz, colorSchemeLight, colorSchemeDark } from 'ag-grid-community';
import { AG_GRID_LOCALE_RU } from './ag-grid-locale-ru.js';
import { provideGlobalGridOptions } from 'ag-grid-community';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community';

ModuleRegistry.registerModules([AllCommunityModule]);

const colorAGLight = themeQuartz.withPart(colorSchemeLight)
    .withParams({
        backgroundColor: "rgba(255, 255, 255, 0.3)",
        headerBackgroundColor: "rgb(233, 235, 236, 0.3)",
    });

const colorAGDark = themeQuartz.withPart(colorSchemeDark)
    .withParams({
        backgroundColor: "rgba(17, 26, 34, 0.3)",
        headerBackgroundColor: "rgb(40, 48, 56, 0.3)",
    });

export default {
    name: 'main-app',
    data() {
        return {
            userBlock: true,
            localeOptions: this.localeSelect(),
            tabValue: 0,
            version: this.$t('message.nodata'),
            user: this.$t('message.nodata'),
            userID: this.$t('message.nodata'),
            showModal: false,
            showPreloader: true,
            curPassword: "",
            newPassword: "",
            newPassword2: "",
            isError1: false,
            isError2: false,
            isError3: false,
            userRoleColor: "#0f1a23ff",
            navBarColor: '#70b8e0ff',
            switchValue: false,
            reliase: this.versionApp,
            logo02
        }
    },
    methods: {
        localeSelect() {
            let vm = this;
            this.localeOptions = [];
            if (this.$i18n.availableLocales.length > 0) {
                this.$i18n.availableLocales.forEach(function(item) {
                    vm.localeOptions.push(item);
                });

            }
            return vm.localeOptions;
        },
        openModalProfile() {
            this.isError1 = false;
            this.isError2 = false;
            this.isError3 = false;
            this.curPassword = "";
            this.newPassword = "";
            this.newPassword2 = "";
        },
        exit() {
            $.ajax({
                url: "/api/v1/exit?" + Math.random(),
                type: "GET",
                success: function () {
                    router.go("/");
                    return true;
                }
            });
        },
        versionInfo() {
            let vm = this;
            $.ajax({
                url: "/api/v1/version?" + Math.random(),
                type: "GET",
                dataType: "json",
                success: function (data) {
                    vm.version = data.Name;
                    return true;
                }
            });
        },
        userInfo() {
            let vm = this;
            $.ajax({
                url: "/api/v1/userinfo?" + Math.random(),
                type: "GET",
                dataType: "json",
                success: function (data) {
                    vm.user = data.lname + " " + data.fname;
                    vm.userID = data.userid;
                    if (data.role == "superadmin") {
                        vm.userRoleColor = "danger";
                        vm.userBlock = false;
                    } else if (data.role == "admin") {
                        vm.userRoleColor = "warning";
                        vm.userBlock = false;
                    } else {
                        vm.userRoleColor = "primary";
                        vm.userBlock = true;
                    }
                    return true;
                }
            });
        },
        putPassword() {
            let vm = this;
            if (vm.CurPassword != "" && vm.NewPassword != "" && vm.newPassword2 != "") {
                if (this.newPassword != this.curPassword) {
                    if (this.newPassword == this.newPassword2) {
                        if ((this.newPassword).length >= 8) {
                            if (confirm(vm.$t('password.confirm'))) {
                                var dataPut = {
                                    'curpassword': sha256(vm.curPassword),
                                    'newpassword': sha256(vm.newPassword),
                                }
                                $.ajax({
                                    url: "/api/v1/password",
                                    type: "PUT",
                                    dataType: "json",
                                    data: JSON.stringify(dataPut),
                                    statusCode: {
                                        200: function() {
                                            vm.$vaToast.init({ message: vm.$t('password.message01'), color: 'primary' });
                                            vm.showModal = false;
                                            return true;
                                        },
                                        404: function() {
                                            vm.$vaToast.init({ message: vm.$t('password.message02'), color: 'danger' });
                                            vm.showModal = true;
                                            return true;
                                        },
                                        500: function() {
                                            vm.$vaToast.init({ message: vm.$t('password.message03'), color: 'danger' });
                                            vm.showModal = true;
                                            return true;
                                        },
                                    },
                                });
                            }
                        } else {
                            vm.$vaToast.init({ message: vm.$t('password.message04'), color: 'danger' });
                            vm.showModal = true;
                            vm.isError1 = true;
                        }
                    } else {
                        vm.$vaToast.init({ message: vm.$t('password.message05'), color: 'danger' });
                        vm.showModal = true;
                        vm.isError2 = true;
                        vm.isError3 = true;
                    }
                } else {
                    vm.$vaToast.init({ message: vm.$t('password.message06'), color: 'danger' });
                    vm.showModal = true;
                    vm.isError1 = true;
                    vm.isError2 = true;
                }
            } else {
                vm.$vaToast.init({ message: vm.$t('password.message07'), color: 'warning' });
                vm.showModal = true;
                vm.isError1 = true;
                vm.isError2 = true;
                vm.isError3 = true;
            }
        },
        /* Switching color theme  */
        switchMode() {
            let vm = this;

            const { setColors } = useColors();
            /* Светлая тема */
            const colorLight = {
                primary: "#1f303eff",
                secondary: "#36536b",
                success: "#66be33",
                info: "#3eaaf8",
                danger: "#f34030",
                warning: "#ffd952",
                backgroundPrimary: "#ffffffff",
                backgroundSecondary: "#ffffffff",
                backgroundElement: "#1f303eff",
                backgroundBorder: "#3d4c58",
                textPrimary: "#1f303eff",
                textInverted: "#ffffffff",
                shadow: "rgba(1, 1, 1, 0.30)",
                focus: "#1f303eff",
                transparent: "rgba(0, 0, 0, 0)",
                backgroundLanding: "#070d14",
                backgroundLandingBorder: "rgba(43, 49, 56, 0.8)"
            };

            /* Тёмная тема */
            const colorDark = {
                primary: "#ffffff",
                secondary: "#A0A0A0",
                success: "#66be33",
                info: "#3eaaf8",
                danger: "#f34030",
                warning: "#ffd952",
                backgroundPrimary: "#0f1a23ff",
                backgroundSecondary: "#1f303eff",
                backgroundElement: "#1f303eff",
                backgroundBorder: "#ffffff",
                textPrimary: "#ffffff",
                textInverted: "#1f303eff",
                shadow: "rgba(255, 255, 255, 0.12)",
                focus: "#ffffff",
                transparent: "rgba(0, 0, 0, 0)",
                backgroundLanding: "#070d14",
                backgroundLandingBorder: "rgba(43, 49, 56, 0.8)"
            };

            vm.$store.commit('changemode', vm.switchValue);
            if (vm.switchValue == false) {
                vm.navBarColor = '#70b8e0ff';
                setColors(colorLight); // Globally set colors for a bright theme
                $('.body').css('background-image', 'var(--bg-svg)');
                $('.up').css('color', '#1f303eff');

                if (this.$i18n.locale == "ru") {
                    provideGlobalGridOptions({
                        theme: colorAGLight,
                        localeText: AG_GRID_LOCALE_RU,
                    });
                } else {
                    provideGlobalGridOptions({
                        theme: colorAGLight,
                        localeText: null,
                    });
                }
            } else {
                vm.navBarColor = '#1f303eff';
                setColors(colorDark); // Globally set colors for a dark theme
                $('.body').css('background-image', 'var(--bg-svg-dark)');
                $('.up').css('color', '#ffffff');
                $('*').css('--va-1-color-computed', '#ffffff');

                if (this.$i18n.locale == "ru") {
                    provideGlobalGridOptions({
                        theme: colorAGDark,
                        localeText: AG_GRID_LOCALE_RU,
                    });
                } else {
                    provideGlobalGridOptions({
                        theme: colorAGDark,
                        localeText: null,
                    });
                }
            }
            console.log("Тёмный режим: " + vm.$store.state.mode);
        },
        localeUpdate() {
            this.$store.commit('changelocale', this.$i18n.locale);
            console.log("Change locale", this.$i18n.locale);

            if (this.$i18n.locale == "ru") {
                if (this.switchValue == false) {
                    provideGlobalGridOptions({
                        localeText: AG_GRID_LOCALE_RU,
                        theme: colorAGLight,
                    });
                } else {
                    provideGlobalGridOptions({
                        localeText: AG_GRID_LOCALE_RU,
                        theme: colorAGDark,
                    });
                }
            } else {
                if (this.switchValue == false) {
                    provideGlobalGridOptions({
                        localeText: null,
                        theme: colorAGLight,
                    });
                } else {
                    provideGlobalGridOptions({
                        localeText: null,
                        theme: colorAGDark,
                    });
                }
            }
        }
    },
    beforeCreate() {
        const colorAGLight = themeQuartz.withPart(colorSchemeLight)
            .withParams({
                backgroundColor: "rgba(255, 255, 255, 0.3)",
                headerBackgroundColor: "rgb(233, 235, 236, 0.3)",
                });
        if (this.$i18n.locale == "ru") {
            provideGlobalGridOptions({
                localeText: AG_GRID_LOCALE_RU,
                theme: colorAGLight,
            });
        } else {
            provideGlobalGridOptions({
                localeText: null,
                theme: colorAGLight,
            });
        }
    }
};
</script>

<style>
* {
    --va-navbar-z-index: 3;
    --va-card-box-shadow: none;
    --va-card-outlined-border: 1px solid var(--va-background-element);
    --va-inner-loading-position: relative;

    --va-toast-width: 430px;

    font-family: Arial, Helvetica, sans-serif;
	font-size: 10pt;
}
.va-dropdown__content {
    z-index: 5 !important; /* Fix z-index for form elements in modal windows */
}
#preloader {
    background-color: #fff;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 999
}
#preloader div {
    background: var(--va-primary);
    width: 350px;
    height: 40px;
    line-height: 40px;
    border-radius: 8px;
    font-family: arial;
    font-size: 15px;
    color: #fff;
    text-align: center;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
    position: fixed;
    z-index: 999;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    margin: auto
}
.mb-4 {
    width: 100%;
    min-width: 400px;
}
.main {
    height: 100vh;
    width: 100%;
    min-width: 800px;
}
.up {
    width: 100%;
    height: auto;
    min-width: 800px;
    align-items: center;
    position: sticky;
    top: 0;
    padding: 5px;
}
.down {
    color: var(--va-primary);
}
.body {
    --bg-svg: url(assets/bg02.svg);
    --bg-svg-dark: url(assets/bg02d.svg);
    background-image: var(--bg-svg);
    background-position: center;
    background-repeat: no-repeat;
    background-size: cover;
    background-attachment: fixed;
    color: #000000;
    width: 100%;
    height: auto;
}
div.user:hover {
    cursor: pointer;
}
.va-navbar__left {
    display: flex;
    align-items: center;
}
.va-navbar__right {
    display: flex;
    align-items: center;
}
div.menu:hover {
    cursor: pointer;
    color: var(--va-warning);
}
div.version {
    font-family: "Roboto Condensed", sans-serif;
	font-size: 24pt;
    cursor: default;
    cursor: pointer;
}
div.logo:hover {
    cursor: pointer;
}
.badge {
    margin-top: 3px;
    font-family: 'Play', sans-serif;
    cursor: default;
    color: var(--va-primary);
    cursor: pointer;
}
div.content {
    color: var(--va-primary);
}
div.test {
    overflow: auto;
}
.layout {
    margin: unset;
    max-width: 100%
}
.va-breadcrumbs__separator {
    padding-left: 0.5rem;
    padding-right: 0.5rem;
}

/* Input adjustment for 1.8.1 */
.va-input-wrapper {
    display: flex; 
    width: 100%;
}
</style>