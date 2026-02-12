export const emptyChartOptions = {
  title: "",
  axes: {
    left: {
      mapsTo: "value",
    },
    bottom: {
      scaleType: "labels",
      mapsTo: "key",
    },
  },
  data: {
    loading: true,
  },
  height: "30rem",
};
