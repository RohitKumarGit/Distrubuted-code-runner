const express = require('express');
const mongoose = require('mongoose');
const cors = require('cors');
const { default: axios } = require('axios');
const { MongoClient, ServerApiVersion } = require('mongodb');

const app = express();
const port =  process.env.PORT || 8085 
const DATABASE_NAME = process.env.DATABASE_NAME || 'code_scheduler'
const MONGODB_URI =

  "mongodb://root:example@localhost:27017";
const MASTER_BASE_URL =  process.env.MASTER_BASE_URL || 'http://localhost:8080' 
// Middleware
app.use(cors());
app.use(express.json());

// MongoDB connection
const client = new MongoClient(MONGODB_URI, {
  serverApi: {
    version: ServerApiVersion.v1,
    strict: true,
    deprecationErrors: true,
  },
});



// Routes
app.get('/jobs', async (req, res) => {
  // sort with newest job first
  const jobs = await client.db(DATABASE_NAME).collection('jobs').find().sort({ _id: -1 }).toArray();
  console.log(jobs)
  res.json(jobs);
});
app.get("/status",async (req,res)=>{
  try {
    const { data } = await axios.get(`${MASTER_BASE_URL}/status`);
    res.json({master_online:true,...data});
  } catch (error) {
    res.json({ master_online: false, workers:[] });
  }
  
})
app.post('/submit', async (req, res) => {
  const { python_code } = req.body;
  try {
    const response = await fetch(`${MASTER_BASE_URL}/submit`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ python_code }),
    });
    
    res.send(response);
  } catch (error) {
    console.error('Error submitting job:', error);
    res.status(500).json({ message: 'Error submitting job' });
  }
});
const listen = async function(){
  await client.connect();
  app.listen(port, () => {
    console.log(`Server is running at http://localhost:${port}`);
  });
}
listen()