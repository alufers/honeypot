<script lang="ts">
  import AttacksList from "./AttacksList.svelte";
  import CountriesTopList from "./CountriesTopList.svelte";
  import PasswordsTopList from "./PasswordsTopList.svelte";
  import ProtocolList from "./ProtocolList.svelte";

  export let appVersion: string = process.env.HONEYPOT_VERSION;
</script>

<main>
  <h1>Honeypot</h1>
  <div class="version">version: {appVersion}</div>
  <div class="stats-columns">
    <div>
      <ProtocolList />
    </div>
    <div>
      <CountriesTopList />
    </div>
    <div>
      <PasswordsTopList />
    </div>
  </div>
  <AttacksList />
</main>

<style lang="scss">
  .version {
    font-size: 0.8em;
    color: #999;
    margin-bottom: 1.5rem;
  }
  h1 {
    margin-block-end: 0;
  }
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
    @media (max-width: 900px) {
      flex-direction: column;
      &> div {
        margin-bottom: 1rem;
      }
    }
  }
</style>
