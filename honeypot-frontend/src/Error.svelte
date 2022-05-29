<script lang="ts">
    export let error: {
        message: string;
    }
</script>
<div>
    {#if error}
        {error.message}
    {/if}
</div>

