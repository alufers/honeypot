<script lang="ts">
  import { fetchAPI } from "./api";
  import DataTable from "./DataTable.svelte";
  import type { IColumn } from "./DataTable.svelte";
  import Error from "./Error.svelte";
  import Loader from "./Loader.svelte";

  const countriesPromise = fetchAPI(
    "GET",
    "/api/attacks/stats/by-country"
  ).then((countries) => {
    let sum = countries.reduce((sum, country) => {
      return sum + country.count;
    }, 0);
    return countries.map((c) => ({
      ...c,
      percentage: ((c.count / sum) * 100).toFixed(1) + "%",
    }));
  });
  const columns: IColumn[] = [
    {
      key: "country_code",
      label: "",
      type: "flag",
    },
    {
      key: "country",
      label: "Country",
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
  Top countries:

  {#await countriesPromise}
    <Loader />
  {:then countries}
    <DataTable {columns} data={countries} limit={10} />
  {:catch error}
    <Error {error} />
  {/await}
</div>
