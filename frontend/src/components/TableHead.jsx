import React from "react";

function TableHead() {
    return <thead className="text-center">
    <tr>
        <th className="col-1">IP</th>
        <th className="col-1">Время пинга</th>
        <th className="col-1">Последняя успешная попытка</th>
    </tr>
    </thead>;
}

export default TableHead;