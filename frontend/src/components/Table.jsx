import React from "react";
import Row from "./Row";

function Table({rows}) {
    if (!Array.isArray(rows) || rows.length === 0)  {
        return;
    }
    return <table className="table table-striped table-bordered table-hover">
        <thead className="text-center">
            <tr>
                <th className="col-1">IP</th>
                <th className="col-1">Время пинга</th>
                <th className="col-1">Последняя успешная попытка</th>
            </tr>
        </thead>
        <tbody className="text-center table-group-divider">
            {rows.map((row, index) => (
                <Row key={index} {...row}/>
            )) }
        </tbody>
    </table>
}

export default Table;