<template>
    <div>
        <va-button preset="plain" icon="edit" v-bind:disabled="userBlock" v-on:click="rowClickEdit(params.node.rowIndex)" />
        <va-button preset="plain" icon="delete" v-bind:disabled="userBlock" v-on:click="rowClickRemove(params.node.rowIndex)" />
    </div>
</template>

<script>
    import $ from 'jquery'
    export default {
        props: {
            params: {
                type: Object,
                required: true,
            },
        },
        methods: {
            rowClickEdit(value) {
                this.params.rowClickEdit(value);
            },
            rowClickRemove(value) {
                this.params.rowClickRemove(value);
            },
        }
    };
</script>