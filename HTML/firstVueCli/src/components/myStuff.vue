<template>
  <div id="myStuff">
    <h2>TASK ONE: Add Value:</h2>
    <p>{{value}}</p>
    <el-button type="primary" v-on:click="addOne" style="width: 15%">+1s</el-button>
    <h2>TASK TWO: Count Word:</h2>
    <el-input v-model="input" placeholder="请输入内容"></el-input>
    <p>has Word <span>{{input.length}}</span> Words</p>
    <h2>TASK Three: Check Validate:</h2>
    <el-input type ="text" v-model="cellPhone"  @input="checkValue" > </el-input>
    <span> {{success}}</span>
  </div>
</template>

<script>
export default {
  name: 'myStuff',
  data () {
    return {input: '', cellPhone: '', success: 'NOT A VALIDATE PHONE NUMBER', value: 0}
  },
  methods: {
    checkValue: function () {
      var format = /^1\d{10}$/
      if (format.test(this.cellPhone)) {
        this.success = 'CORRECT!'
        return 1
      }
      this.success = 'NOT A VALIDATE PHONE NUMBER!'
      return 0
    },
    addOne: function () {
      this.value = this.value + 1
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
