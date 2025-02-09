import React, { useState, useEffect } from "react";
import Table from "./components/Table";

function App() {
  const [data, setData] = useState([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch("http://localhost:8080/pings")
        const result = await response.json()
        setData(result)
      } catch (error) {
        console.error("Ошибка при загрузке данных: ", error);
      }
    }

    fetchData();
    const interval = setInterval(fetchData, 30000);
    return () => clearInterval(interval);

  }, []);


  return <div className="App">
    <Table rows={data}/>
  </div>
}

export default App;
