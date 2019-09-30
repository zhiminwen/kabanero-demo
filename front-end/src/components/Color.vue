<template>
<v-item-group> 
    <v-container>
      <v-row>
        <v-col
          v-for="c in colors"
          :key="c.index"
          cols="12"
          md="4"
        >
          <v-item>
            <v-card
              :color="c.color"
              class="d-flex align-center"
              dark
              height="150"
            >
            <div class="display-3 flex-grow-1 text-center">{{ c.version }}</div>
            </v-card>
          </v-item>
        </v-col>
      </v-row>
    </v-container>
  </v-item-group>
  
  
</template>

<script>
import axios from 'axios';
export default { 
  data() {
    return {
      colors: [],
    }
  },
  mounted(){
    for (let index = 0; index < 10; index++) {
      axios.post("/api/getcolor").then( color => {
        // use push to make the data responsive other wise, its not working
        this.colors.push({
          index: index,
          color: color.data.color,
          version: color.data.version
        })
        console.log(this.colors)
      }).catch(err=>console.error(err))
    }
    window.mydata=this.colors
  },
  methods: {
    
  }
}
</script>

<style>

</style>