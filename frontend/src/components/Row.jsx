import React from "react";

function Row({ip, duration, time_attempt}) {
    const durationMs = duration / 1000;
    const formattedDuration = durationMs.toFixed(3);

    const formattedTime = new Date(time_attempt).toLocaleString();
    return <tr>
        <td>{ip}</td>
        <td>{formattedDuration} ms</td>
        <td>{formattedTime}</td>
    </tr>
}

export default Row;