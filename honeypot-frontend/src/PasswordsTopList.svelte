<script lang="ts">
    import { fetchAPI } from "./api";
    import DataTable from "./DataTable.svelte";
    import type { IColumn } from "./DataTable.svelte";
    import Error from "./Error.svelte";
    import Loader from "./Loader.svelte";
  
    const passwordsPromise = fetchAPI(
      "GET",
      "/api/credentials/stats/passwords"
    ).then((countries) => {
     
      return countries.map((c) => ({
        ...c,
        percentage: c.percentage.toFixed(1) + "%",
      }));
    });
    const columns: IColumn[] = [
      {
        key: "password",
        label: "Password",
       
      },
      {
        key: "count",
        label: "Count",
      },
      {
        key: "percentage",
        label: "%",
      },
    ];
  </script>
  
  <div>
    Top attempted passwords:
  
    {#await passwordsPromise}
      <Loader />
    {:then passwords}
      <DataTable {columns} data={passwords} limit={10} />
    {:catch error}
      <Error {error} />
    {/await}
  </div>
  