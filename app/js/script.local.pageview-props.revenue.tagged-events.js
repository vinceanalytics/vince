!function(){"use strict";var u=window.location,o=window.document,s=o.currentScript,l=s.getAttribute("data-api")||new URL(s.src).origin+"/api/event";function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return r=t,(n="localStorage flag")&&console.warn("Ignoring Event: "+n),void(r&&r.callback&&r.callback())}catch(e){}var n={},r=(n.n=e,n.u=u.href,n.d=s.getAttribute("data-domain"),n.r=o.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),s.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),a=n.p||{},i=(r.forEach(function(e){var t=e.replace("event-",""),e=s.getAttribute(e);a[t]=a[t]||e}),n.p=a,new XMLHttpRequest);i.open("POST",l,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n,r=0;r<t.length;r++)e.apply(this,t[r]);function a(){n!==u.pathname&&(n=u.pathname,e("pageview"))}var i,p=window.history;function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}p.pushState&&(i=p.pushState,p.pushState=function(){i.apply(this,arguments),a()},window.addEventListener("popstate",a)),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){n||"visible"!==o.visibilityState||a()}):a();var f=1;function v(e){"auxclick"===e.type&&e.button!==f||((e=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target))&&e.href&&e.href.split("?")[0],function e(t,n){if(!t||g<n)return!1;if(h(t))return!0;return e(t.parentNode,n+1)}(e,0))}function d(e,t,n){var r,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((r={props:n.props}).revenue=n.revenue,plausible(n.name,r)):((r={props:n.props,callback:i}).revenue=n.revenue,plausible(n.name,r),setTimeout(i,5e3),e.preventDefault())}function m(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},n=e&&e.classList;if(n)for(var r=0;r<n.length;r++){var a,i,u=n.item(r),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(a=o[1],i=o[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(a=o[1],i=o[3],t.revenue[a]=i)}return t}o.addEventListener("click",v),o.addEventListener("auxclick",v);var g=3;function b(e){if("auxclick"!==e.type||e.button===f){for(var t,n,r,a,i=e.target,u=0;u<=g&&i;u++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;c(i)&&(t=i),h(i)&&(n=i),i=i.parentNode}n&&(a=m(n),t?(a.props.url=t.href,d(e,t,a)):((e={}).props=a.props,e.revenue=a.revenue,plausible(a.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,n=e.target,r=m(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),(e={props:r.props,callback:a}).revenue=r.revenue,plausible(r.name,e))}),o.addEventListener("click",b),o.addEventListener("auxclick",b)}();