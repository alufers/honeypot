<script lang="ts">
  import AttacksList from "./AttacksList.svelte";
  import CountriesTopList from "./CountriesTopList.svelte";
  import ProtocolList from "./ProtocolList.svelte";

  export let name: string;
</script>

<main>
  <h1>Honeypot</h1>
  <div class="stats-columns">
    <div>
      <ProtocolList />
    </div>
    <div>
      <CountriesTopList />
    </div>
  </div>
  <AttacksList />
</main>

<style lang="scss">
  main {
    padding: 1em;
    max-width: 1200px;
    margin: 0 auto;
  }
  .stats-columns {
    display: flex;
    flex-direction: row;
    > div {
      flex-grow: 1;
    }
  }
</style>
