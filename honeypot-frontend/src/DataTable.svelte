<script lang="ts" context="module">
  export interface IColumn {
    label: string;
    key: string;
    type?: "text" | "flag";
  }
</script>

<script lang="ts">
  import Button from "./Button.svelte";

  import Flag from "./Flag.svelte";

  export let columns: IColumn[];
  export let data: any[];
  export let limit: number | null = null;
  let expanded = false;
  $: dataLimited = data.filter((_, i) => {
    if (limit === null || expanded) {
      return true;
    }
    return i < limit;
  });
</script>

<table>
  <thead>
    <tr>
      {#each columns as column}
        <th>{column.label}</th>
      {/each}
    </tr>
  </thead>
  <tbody>
    {#each dataLimited as row}
      <tr>
        {#each columns as column}
          <td>
            {#if column.type === "flag"}
              <Flag country={row[column.key]} />
            {/if}
            {#if column.type === "text" || !column.type}
              {row[column.key]}
            {/if}
          </td>
        {/each}
      </tr>
    {/each}
  </tbody>
  {#if limit !== null && data.length > limit}
    <tfoot>
      <tr>
        <td colspan={columns.length}>
          <Button
            on:click={() => {
                
              expanded = !expanded;
            }}
          >
            {expanded
              ? "Show less"
              : "Show more (" + (data.length - limit) + " more)"}
          </Button>
        </td>
      </tr>
    </tfoot>
  {/if}
</table>

<style>
  table {
    border-collapse: collapse;
  }
  th {
    padding: 0.2em;
    text-align: left;
  }
  td {
    padding-top: 0.2rem;
    padding-right: 0.2rem;
  }
  th {
    border-bottom: 1px solid #eee;
  }
</style>
