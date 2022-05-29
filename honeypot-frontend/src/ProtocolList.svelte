<script lang="ts">
import { fetchAPI } from "./api";
import Error from "./Error.svelte";
import Loader from "./Loader.svelte";

const protocolsPromise = fetchAPI("GET", "/api/protocols");
</script>

<div>
    Active protocols:
    <ul>
        {#await protocolsPromise}
            <Loader />
        {:then protocols}
            {#each protocols  as protocol}
                <li>
                    {protocol}
                </li>
            {/each}
        {:catch error}
            <Error error={error} />
        {/await}
    </ul>
</div>
