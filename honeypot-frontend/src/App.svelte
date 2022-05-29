<script lang="ts">
import AttacksList from "./AttacksList.svelte";
import ProtocolList from "./ProtocolList.svelte";

	export let name: string;
</script>

<main>
	<h1>Honeypot</h1>
	<ProtocolList></ProtocolList>
	<AttacksList></AttacksList>
</main>

<style>
	main {
		
		padding: 1em;
		max-width: 1200px;
		margin: 0 auto;
	}


</style>
