!function(){"use strict";var u=window.location,l=window.document,s=l.currentScript,c=s.getAttribute("data-api")||new URL(s.src).origin+"/api/event";function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(u.hostname)||"file:"===u.protocol)return p("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return p(null,t);try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var r=s&&s.getAttribute("data-include"),n=s&&s.getAttribute("data-exclude");if("pageview"===e){r=!r||r.split(",").some(a),n=n&&n.split(",").some(a);if(!r||n)return p("exclusion rule",t)}function a(e){var t=u.pathname;return(t+=u.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var r={},n=(r.n=e,r.u=t&&t.u?t.u:u.href,r.d=s.getAttribute("data-domain"),r.r=l.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),s.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=r.p||{},o=(n.forEach(function(e){var t=e.replace("event-",""),e=s.getAttribute(e);i[t]=i[t]||e}),r.p=i,r.h=1,new XMLHttpRequest);o.open("POST",c,!0),o.setRequestHeader("Content-Type","text/plain"),o.send(JSON.stringify(r)),o.onreadystatechange=function(){4===o.readyState&&t&&t.callback&&t.callback({status:o.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function f(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var v=1;function n(e){var t,r;if("auxclick"!==e.type||e.button===v)return(t=function(e){for(;e&&(void 0===e.tagName||!f(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||g<r)return!1;if(h(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(r=t)&&r.href&&r.host&&r.host!==u.host?d(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function d(e,t,r){var n,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((n={props:r.props}).revenue=r.revenue,plausible(r.name,n)):((n={props:r.props,callback:i}).revenue=r.revenue,plausible(r.name,n),setTimeout(i,5e3),e.preventDefault())}function m(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},r=e&&e.classList;if(r)for(var n=0;n<r.length;n++){var a,i,o=r.item(n),u=o.match(/plausible-event-(.+)(=|--)(.+)/),u=(u&&(a=u[1],i=u[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i),o.match(/plausible-revenue-(.+)(=|--)(.+)/));u&&(a=u[1],i=u[3],t.revenue[a]=i)}return t}l.addEventListener("click",n),l.addEventListener("auxclick",n);var g=3;function a(e){if("auxclick"!==e.type||e.button===v){for(var t,r,n,a,i=e.target,o=0;o<=g&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;f(i)&&(t=i),h(i)&&(r=i),i=i.parentNode}r&&(a=m(r),t?(a.props.url=t.href,d(e,t,a)):((e={}).props=a.props,e.revenue=a.revenue,plausible(a.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}l.addEventListener("submit",function(e){var t,r=e.target,n=m(r);function a(){t||(t=!0,r.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),(e={props:n.props,callback:a}).revenue=n.revenue,plausible(n.name,e))}),l.addEventListener("click",a),l.addEventListener("auxclick",a)}();