import axios from 'axios'
import Express from 'express'

const app = Express()
const maxRetries = 3

app.get('/ping', async (req, res) => {
  const { limit } = req.query
  console.log(limit)
  let response = []
  for (let i = 0; i < limit; i++) {
    const data = await getSomething(i, axiosTimeout)
    response.push(data)
  }
  res.json(response)
})

app.get('/pong', async (req, res) => {
  const { limit } = req.query
  console.log(limit)
  //   let promises = []
  //   for (let i = 0; i < limit; i++) {
  //     promises.push(getSomethingWithRetry(i, maxRetries))
  //   }
  const axiosTimeout = 90000
  const promises = Array.from({ length: limit }, (_, i) => getSomething(i, axiosTimeout))

  try {
    const response = await Promise.all(promises)
    res.json(response)
  } catch (error) {
    console.error(error.message)
    res.status(500).json({ error: 'An error occurred' })
  }
})

app.listen(3000, () => {
  console.log('Server running on port 3000')
})

async function getSomething(id, timeout) {
  try {
    for (let retries = 0; retries < maxRetries; retries++) {
      try {
        const { data } = await axios.get(`http://localhost:8080/api/v1/get_something?id=${id}`, { timeout })
        return data
      } catch (error) {
        if (error.code === 'ECONNRESET') {
          // Handle the ECONNRESET error, possibly by retrying the request
          console.error(`ECONNRESET error for request id ${id}. Retrying...`)
        } else {
          console.error(`Request for id ${id} failed: ${error.message}`)
          break // Break out of the retry loop for other error types
        }
      }
    }
  } catch (error) {
    // Handle request timeout or other errors
    console.error(`Request for id ${id} failed: ${error.message}`)
    return { id, error: 'Request failed' }
  }
}
async function getSomethingWithRetry(id, maxRetries) {
  try {
    for (let retries = 0; retries < maxRetries; retries++) {
      try {
        const { data } = await axios.get(`http://localhost:8080/api/v1/get_something?id=${id}`)
        return data
      } catch (error) {
        if (error.code === 'ECONNRESET') {
          // Handle the ECONNRESET error, possibly by retrying the request
          console.error(`ECONNRESET error for request id ${id}. Retrying...`)
        } else {
          console.error(`Request for id ${id} failed: ${error.message}`)
          break // Break out of the retry loop for other error types
        }
      }
    }

    return { id, error: 'Max retries reached' }
  } catch (error) {
    console.error(`Request for id ${id} failed: ${error.message}`)
    return { id, error: 'An error occurred' }
  }
}
