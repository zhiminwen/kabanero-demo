const express = require('express');
const path = require('path');
const bodyParser = require("body-parser");
const axios = require("axios");

const app = express();
app.use(bodyParser.json());
app.use(express.static(path.resolve(__dirname, "..", "dist")));

// Always return the main index.html, so react-router render the route in the client
app.get("*", (req, res) => {
  res.sendFile(path.resolve(__dirname, "..", "dist", "index.html"));
});

var port = process.env.APP_PORT || 5000;
app.listen(port);

var restServerBase = process.env.APP_REST_SERVER || "http://localhost:9691";

app.post("/api/getcolor", (req, res) => {
  axios.post(restServerBase + "/getcolor").then(color => {
    console.log(color)
    res.send({
      color: color.data.Color,
      version: color.data.Version
    })
  }).catch(err=> {
    console.log(err)
    res.status(500).send('Something broke!')
  })
})

// module.exports = app;

