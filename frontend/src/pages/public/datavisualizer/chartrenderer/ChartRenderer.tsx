import React, { useEffect, useState } from "react";
import API from "../../helpers/api";
import PivotTableUI from "react-pivottable/PivotTableUI";
import TableRenderers from "react-pivottable/TableRenderers";
import Plot from "react-plotly.js";
import createPlotlyRenderers from "react-pivottable/PlotlyRenderers";

import "react-pivottable/pivottable.css";
// import {useToast} from "../../../../components/notifications/toast/useToast.ts";

const ChartRenderer = ({
  queryParams,
  onSaveLoadedData,
  onSavePivotData,
  loadedData,
  pivotData,
  periods,
}) => {
  const PlotlyRenderers = createPlotlyRenderers(Plot);
  const [loading, setLoading] = useState(false);
  const [chartData, setChartData] = useState(loadedData);
  const [pivotTableData, setPivotTableData] = useState(pivotData);
  // const toast = useToast();

  useEffect(() => {
    if (!queryParams) return;

    const fetchChartData = async () => {
      setLoading(true);

      try {
        const response = await API.post("/visualizer/datavalues", queryParams);
        const { status, data } = response;

        if (status === 200) {
          const rows = data.rows || [];
          const mapped_pivot_data =
            rows?.map((item) => ({
              "Age-Sex Disaggregation": item.co,
              District: item.district,
              "Data Element": item.dxName,
              "Facility Name": item.ouName,
              Period: periods.find((period) => period.id === item.pe)?.label,
              Region: item.region,
              SubCounty: item.subCounty,
            })) ?? [];

          setChartData(rows);
          onSaveLoadedData(rows);
          setPivotTableData(mapped_pivot_data);
          onSavePivotData(mapped_pivot_data);
        } else {
          // toast.error(`Failed to fetch data: Status ${status}`);
          setChartData([]);
        }
      } catch (err) {
        // toast.error("Error encountered while fetching data values");
        setChartData([]);
      } finally {
        setLoading(false);
      }
    };

    fetchChartData();
  }, [queryParams]);

  return (
    <>
      {loading ? (
        <div> Loading... </div>
      ) : (
        // <StackedBarChart data={[]} options={emptyChartOptions} />
        <>
          {chartData.length === 0 && !loading && (
            <div className="dv-canvas-placeholder">
              <i className="fas fa-chart-bar"></i>
              <h6>No Data Available for the selection</h6>
            </div>
          )}

          {chartData?.length > 0 && (
            <div className={`dwh-chart-container`}>
              <PivotTableUI
                data={pivotTableData}
                cols={["Data Element"]}
                onChange={(s) => setPivotTableData(s)}
                renderers={{ ...TableRenderers, ...PlotlyRenderers }}
                {...pivotTableData}
              />
            </div>
          )}
        </>
      )}
    </>
  );
};

export default ChartRenderer;
