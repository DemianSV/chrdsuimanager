<template>
    <div class="bredcrumbs">
        <va-breadcrumbs class="mb-4">
            <va-breadcrumbs-item label="Главная страница" to="/" />
            <va-breadcrumbs-item label="Справка" to="/about" disabled />
        </va-breadcrumbs>
        </div>
    <div class="page"></div>
</template>

<script>
    export default {
        name: 'about-app',
        data() {
            return {
            }
        },
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