// src/api.js
import axios from 'axios'
const api = axios.create({
  baseURL: 'http://localhost:8085/api',
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  },
  withCredentials: false
})
export default api


// const BASE_URL = 'http://localhost:8080'

// export const api = {
//   // Get all rioters
//   async getRioters() {
//     try {
//       const response = await fetch(`${BASE_URL}/api/rioters`)
//       if (!response.ok) {
//         throw new Error('Network response was not ok')
//       }
//       const data = await response.json()
//       console.log('Fetched rioters:', data) // Debug log
//       return data
//     } catch (error) {
//       console.error('Error fetching rioters:', error)
//       throw error
//     }
//   },
//   // Add a new rioter
//   async addRioter(rioter) {
//     try {
//       const response = await fetch(`${BASE_URL}/api/rioters`, {
//         method: 'POST',
//         headers: {
//           'Content-Type': 'application/json',
//         },
//         body: JSON.stringify(rioter)
//       })
//       return await response.json()
//     } catch (error) {
//       console.error('Error adding rioter:', error)
//       throw error
//     }
//   }
// }