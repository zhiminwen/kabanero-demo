<template>
  <v-app>
    <v-app-bar app>
      <v-toolbar-title class="headline text-uppercase">
        <span>Kabanero Micro-Servies Deployment</span>
      </v-toolbar-title>
    </v-app-bar>
    <v-content>
      <Color/>
    </v-content>
  </v-app>
</template>

<script>
// import HelloWorld from './components/HelloWorld';
import Color from './components/Color';

export default {
  name: 'App',
  components: {
    // HelloWorld,
    Color,
  },
  data: () => ({
    //
  }),
};
</script>
