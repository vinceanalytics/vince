import { Menu, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import React, { Fragment, useCallback, useEffect } from 'react';
import classNames from 'classnames';
import * as storage from '../../util/storage';
import { isKeyPressed } from '../../keybinding.js';
import { useQueryContext } from '../../query-context.js';
import { useSiteContext } from '../../site-context.js';

const INTERVAL_LABELS = {
  'minute': 'Minutes',
  'hour': 'Hours',
  'date': 'Days',
  'week': 'Weeks',
  'month': 'Months'
}

function validIntervals(site, query) {
  if (query.period === 'custom') {
    if (query.to.diff(query.from, 'days') < 7) {
      return ['date']
    } else if (query.to.diff(query.from, 'months') < 1) {
      return ['date', 'week']
    } else if (query.to.diff(query.from, 'months') < 12) {
      return ['date', 'week', 'month']
    } else {
      return ['week', 'month']
    }
  } else {
    return site.validIntervalsByPeriod[query.period]
  }
}

function getDefaultInterval(query, validIntervals) {
  const defaultByPeriod = {
    'day': 'hour',
    '7d': 'date',
    '6mo': 'month',
    '12mo': 'month',
    'year': 'month'
  }

  if (query.period === 'custom') {
    return defaultForCustomPeriod(query.from, query.to)
  } else {
    return defaultByPeriod[query.period] || validIntervals[0]
  }
}

function defaultForCustomPeriod(from, to) {
  if (to.diff(from, 'days') < 30) {
    return 'date'
  } else if (to.diff(from, 'months') < 6) {
    return 'week'
  } else {
    return 'month'
  }
}

function getStoredInterval(period, domain) {
  return storage.getItem(`interval__${period}__${domain}`)
}

function storeInterval(period, domain, interval) {
  storage.setItem(`interval__${period}__${domain}`, interval)
}

function subscribeKeybinding(element) {
  // eslint-disable-next-line react-hooks/rules-of-hooks
  const handleKeyPress = useCallback((event) => {
    if (isKeyPressed(event, "i")) element.current?.click()
  }, [])

  // eslint-disable-next-line react-hooks/rules-of-hooks
  useEffect(() => {
    document.addEventListener('keydown', handleKeyPress)
    return () => document.removeEventListener('keydown', handleKeyPress)
  }, [handleKeyPress])
}

export const getCurrentInterval = function(site, query) {
  const options = validIntervals(site, query)

  const storedInterval = getStoredInterval(query.period, site.domain)
  const defaultInterval = getDefaultInterval(query, options)

  if (storedInterval && options.includes(storedInterval)) {
    return storedInterval
  } else {
    return defaultInterval
  }
}

export function IntervalPicker({ onIntervalUpdate }) {
  const {query} = useQueryContext();
  const site = useSiteContext();
  if (query.period == 'realtime') return null
  
  // eslint-disable-next-line react-hooks/rules-of-hooks
  const menuElement = React.useRef(null)
  const options = validIntervals(site, query)
  const currentInterval = getCurrentInterval(site, query)

  subscribeKeybinding(menuElement)

  function updateInterval(interval) {
    storeInterval(query.period, site.domain, interval)
    onIntervalUpdate(interval)
  }

  function renderDropdownItem(option) {
    return (
      <Menu.Item onClick={() => updateInterval(option)} key={option} disabled={option == currentInterval}>
        {({ active }) => (
          <span className={classNames({
            'bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-200 cursor-pointer': active,
            'text-gray-700 dark:text-gray-200': !active,
            'font-bold cursor-none select-none': option == currentInterval,
          }, 'block px-4 py-2 text-sm')}>
            {INTERVAL_LABELS[option]}
          </span>
        )}
      </Menu.Item>
    )
  }

  return (
    <Menu as="div" className="relative inline-block pl-2">
      {({ open }) => (
        <>
          <Menu.Button ref={menuElement} className="text-sm inline-flex focus:outline-none text-gray-700 dark:text-gray-300 hover:text-indigo-600 dark:hover:text-indigo-600 items-center">
            {INTERVAL_LABELS[currentInterval]}
            <ChevronDownIcon className="ml-1 h-4 w-4" aria-hidden="true" />
          </Menu.Button>

          <Transition
            as={Fragment}
            show={open}
            enter="transition ease-out duration-100"
            enterFrom="opacity-0 scale-95"
            enterTo="opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="opacity-100 scale-100"
            leaveTo="opacity-0 scale-95">
            <Menu.Items className="py-1 text-left origin-top-right absolute right-0 mt-2 w-56 rounded-md shadow-lg bg-white dark:bg-gray-800 ring-1 ring-black ring-opacity-5 focus:outline-none z-10" static>
              {options.map(renderDropdownItem)}
            </Menu.Items>
          </Transition>
        </>
      )}
    </Menu>
  )
}
