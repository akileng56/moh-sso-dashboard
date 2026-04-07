import {Button, Modal, NumberInput, Select, SelectItem} from "@carbon/react";
import { useMemo, useState } from "react";

import { getAvailablePeriods, getPeriodType, periodType } from "../../Constants.tsx";
import "./period.scss";
import Panel from "../../components/panel/panel.component.tsx";
import {ArrowLeft, ArrowRight} from "@carbon/icons-react";

export default function PeriodModal({ onClose, selected, onSave }) {
  const CURRENT_YEAR = new Date().getFullYear();
  const initialPeriodType = selected?.length > 0 ? getPeriodType(selected[0]?.id) : "Monthly";
  const initialSelectedYear = selected?.length > 0 ? Number(selected[0]?.id?.slice(0,4)) : CURRENT_YEAR;
  const initialPeriods = useMemo(
    () => getAvailablePeriods(initialPeriodType, initialSelectedYear),
    [initialPeriodType, initialSelectedYear],
  );

  const [availablePeriods, setAvailablePeriods] = useState(initialPeriods);
  const [selectedPeriods, setSelectedPeriods] = useState(selected);
  const [selectedPeriodType, setSelectedPeriodType] = useState(initialPeriodType);
  const [selectedYear, setSelectedYear] = useState(initialSelectedYear);

  // const onChangeSelectedPeriod = (event) => {
  //   setSelectedPeriods(event?.selectedItems);
  // };

  const save = () => {
    onSave(selectedPeriods);
    onClose();
  };

  const onChangePeriod = (event) => {
    const paramPeriodType = event?.target?.value;
    setSelectedPeriodType(paramPeriodType);
    const newPeriods = getAvailablePeriods(paramPeriodType, selectedYear);
    setAvailablePeriods(newPeriods);
    setSelectedPeriods([]);
  };

  const onChangeYear = (_event,{ value }) => {
    const paramYear = value;
    setSelectedYear(paramYear);
    const newPeriods = getAvailablePeriods(selectedPeriodType, paramYear);
    setAvailablePeriods(newPeriods);
    setSelectedPeriods([]);
  };

  // const comparePeriodItemsItems = (periodA, periodB) => {
  //   return periodA?.id?.localeCompare(periodB?.id);
  // };

  // const sortPeriodFunction = (periodItems) => {
  //   return [...periodItems]?.sort(comparePeriodItemsItems);
  // };

  // const handleClearTag = (periodId) => {
  //   setSelectedPeriods(selectedPeriods.filter(period => period.id !== periodId));
  // };

  const moveToRight = (selectedItem) => {
    const updatedAvailable = availablePeriods.filter(
        (period) => period?.id !== selectedItem?.id
    );
    setAvailablePeriods(updatedAvailable);
    setSelectedPeriods([...selectedPeriods, selectedItem]);
  };

  const moveToLeft = (selectedItem) => {
    const updatedSelected = selectedPeriods.filter(
        (period) => period.id !== selectedItem.id
    );
    setSelectedPeriods(updatedSelected);
    setAvailablePeriods([...availablePeriods, selectedItem]);
  };

  const moveAllToRight = () => {
    if (availablePeriods.length === 0) return;

    setSelectedPeriods([...selectedPeriods, ...availablePeriods]);
    setAvailablePeriods([]);
  };

  const moveAllToLeft = () => {
    if (selectedPeriods.length === 0) return;
    setAvailablePeriods([...availablePeriods, ...selectedPeriods]);
    setSelectedPeriods([]);
  };

  return (
    <Modal
      open
      size="md"
      preventCloseOnClickOutside={true}
      hasScrollingContent={true}
      modalHeading="Period"
      secondaryButtonText="Hide"
      primaryButtonText="Update"
      onRequestClose={onClose}
      onRequestSubmit={save}
    >
      <div className="row pb-3">
        <div className="col-md-12 d-flex">
          <div className="col-md-6 pe-4">
            <Select
              id={`period-type-select`}
              labelText="Period Type"
              onChange={onChangePeriod}
              value={selectedPeriodType ?? ""}
            >
              <SelectItem text="" value="" />
              {periodType?.map((dataset) => (
                <SelectItem key={dataset?.value} value={dataset?.value} text={dataset?.label} />
              ))}
            </Select>
          </div>
          <div className="col-md-6">
            <NumberInput
              defaultValue={CURRENT_YEAR}
              id="period-year-input"
              invalidText="Input is not a valid year"
              label="Year"
              locale="en"
              max={CURRENT_YEAR}
              min={1900}
              size="md"
              step={1}
              type="number"
              value={selectedYear}
              onChange={onChangeYear}
            />
          </div>
        </div>
      </div>
      <div className="row">
        {/*<div className="mb-2 fw-bold">{selectedPeriodType} Periods</div>*/}

        {/*<div className={`multi-select-period-container`}>*/}
        {/*  <MultiSelect*/}
        {/*    id="period-multiselect-id"*/}
        {/*    label=""*/}
        {/*    titleText="title"*/}
        {/*    onChange={onChangeSelectedPeriod}*/}
        {/*    hideLabel*/}
        {/*    items={availablePeriods}*/}
        {/*    sortItems={sortPeriodFunction}*/}
        {/*    selectedItems={selectedPeriods}*/}
        {/*  />*/}
        {/*  <div className={`selected-period-container`}>*/}
        {/*    {selectedPeriods.map((period) => (*/}
        {/*        <Tag*/}
        {/*            key={period.id}*/}
        {/*            type="blue"*/}
        {/*            filter*/}
        {/*            onClose={() => handleClearTag(period.id)}*/}
        {/*        >*/}
        {/*          {period.label}*/}
        {/*        </Tag>*/}
        {/*    ))}*/}
        {/*  </div>*/}
        {/*</div>*/}

        <div className={`panel-container`}>
          <Panel heading={`Available Periods`}>
            <ul className={`list`}>
              {availablePeriods.map((parameter) => (
                  <li
                      role="menuitem"
                      className={`left-list-item`}
                      key={parameter.label}
                      onClick={() => moveToRight(parameter)}
                  >
                    {parameter.label}
                  </li>
              ))}
            </ul>
          </Panel>
          <div className={`periods-control-container`}>
            <Button
                iconDescription="Move all parameters to the right"
                kind="tertiary"
                hasIconOnly
                renderIcon={ArrowRight}
                onClick={moveAllToRight}
                role="button"
                size="md"
                disabled={availablePeriods.length < 1}
            />
            <Button
                iconDescription="Move all parameters to the left"
                kind="tertiary"
                hasIconOnly
                renderIcon={ArrowLeft}
                onClick={moveAllToLeft}
                role="button"
                size="md"
                disabled={selectedPeriods.length < 1}
            />
          </div>
          <Panel heading="Selected Periods">
            <ul className={`list`}>
              {selectedPeriods.map((parameter) => (
                  <>
                    <li
                        className={`right-list-item`}
                        key={parameter.label}
                        role="menuitem"
                        onClick={() => moveToLeft(parameter)}
                    >
                      {parameter.label}
                    </li>
                  </>
              ))}
            </ul>
          </Panel>
        </div>
      </div>
    </Modal>
  );
}
