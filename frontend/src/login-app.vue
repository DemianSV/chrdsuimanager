<template>
    <div id="preloader" v-if="showPreloader">
        <div>
            {{ $t('login.message05') }}
        </div>
    </div>

    <div class="login">
        <div class="left">
            <!--<div><img :src="require('@/assets/logo01.png')" height="500" align="right"></div>-->
            <div><img :src="logo01" height="450" align="right"></div>
        </div>
        <div class="right">
            <div>
                <div class="row flex layout md12 divFormHeader">Χάρυβδις&nbsp;UI&nbsp;Manager</div>
                <div class="row flex layout md12 divFormHeaderBottom">Charybdis - Opens Source Monitoring System</div>

                <va-form>
                    <div class="row flex layout va-gutter-3 md12"><va-input v-on:focus="this.isError1=false" v-bind:error="isError1" v-model="username" v-on:keydown.enter="login()" :placeholder="$t('login.username01')" :label="$t('login.username02')"></va-input></div>
                    <div class="row flex layout va-gutter-3 md12"><va-input type="password" class="mb-4" v-on:focus="this.isError2=false" v-bind:error="isError2" v-model="password" v-on:keydown.enter="login()" :placeholder="$t('login.password01')" :label="$t('login.password02')"></va-input></div>
                    <div class="row flex layout va-gutter-3 md12"><va-button v-on:click="login()" v-bind:block="isBlock" v-bind:disabled="isButtonDisabled"> {{ $t('login.button') }} </va-button></div>
                </va-form>
            </div>
        </div>
    </div>
</template>

<script>
import sha256 from 'sha256'
import $ from 'jquery'
import router from './router.js';
import logo01 from '@/assets/logo01.svg';

export default {
    name: 'login-app',
    data() {
        return {
            placeholder01: this.$t('login.username02'),
            placeholder02: this.$t('login.password02'),
            error: this.$t('login.message01'),
            errorColor: "",
            username: "",
            password: "",
            isDisabled: false,
            showPreloader: true,
            isBlock: true,
            isError1: false,
            isError2: false,
            isButtonDisabled: false,
            logo01
        }
    },
    methods: {
        login() {
            let vm = this;
            vm.isDisabled = true;
            if( vm.username == '' || vm.password == '') {
                vm.$vaToast.init({ message: vm.$t('login.message02'), color: 'warning' });
                vm.isError1 = true;
                vm.isError2 = true;
            } else {
                vm.isButtonDisabled = true;
                $.post("/auth", {username: vm.username, password: sha256(vm.password)},
                    function() {
                        vm.$vaToast.init({ message: vm.$t('login.message03'), color: 'primary' });
                        vm.isError1 = false;
                        vm.isError2 = false;
                        setTimeout(function() {
                            router.go("/");
                        }, 1000);
                        return true;
                    }
                )
                .fail(
                    function() {
                        vm.$vaToast.init({ message: vm.$t('login.message04'), color: 'danger' });
                        vm.isButtonDisabled = false;
                        vm.isError1 = true;
                        vm.isError2 = true;
                        return true;
                    }
                );
            }
        }
    }
};
</script>

<style>
* {
    font-family: Arial, Helvetica, sans-serif;
	font-size: 10pt;
}

div.login {
    background: url(assets/bg01.svg);
    background-position: center;
    background-repeat: no-repeat;
    background-size: cover;
    height: 100vh;
    width: 100vw;
    align-items: center;
}
div.left {
    float: left;
    width: 50vw;
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: right;
    padding: 1.5rem;
}

div.right {
    float: left;
    width: 50vw;
    height: 100vh;
    display: flex;
    align-items: center;
    padding: 1.5rem;
}
div.divFormHeader {
    font-family: "Roboto Condensed", sans-serif;
    font-optical-sizing: auto;
    font-weight: 500;
    font-size: 48px;
    font-style: normal;
    color: #1f303eff;
}
div.divFormHeaderBottom {
    font-family: "Roboto Condensed", sans-serif;
    font-size: 14pt;
    margin-top: 3px;
    margin-bottom: 30px;
    color: #1f303eff;
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

.va-input-wrapper {
    display: flex;
    width: 100%;
}
</style>