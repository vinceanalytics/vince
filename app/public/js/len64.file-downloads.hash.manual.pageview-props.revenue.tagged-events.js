!function(){"use strict";var i=window.location,o=window.document,u=o.currentScript,l=u.getAttribute("data-api")||new URL(u.src).origin+"/api/event";function p(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return p("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return p(null,t);try{if("true"===window.localStorage.plausible_ignore)return p("localStorage flag",t)}catch(e){}var r={},e=(r.n=e,r.u=t&&t.u?t.u:i.href,r.d=u.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),t&&t.revenue&&(r.$=t.revenue),u.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),n=r.p||{},a=(e.forEach(function(e){var t=e.replace("event-",""),e=u.getAttribute(e);n[t]=n[t]||e}),r.p=n,r.h=1,new XMLHttpRequest);a.open("POST",l,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(r)),a.onreadystatechange=function(){4===a.readyState&&t&&t.callback&&t.callback({status:a.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function s(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var c=1;function n(e){var t,r,n,a;if("auxclick"!==e.type||e.button===c)return t=function(e){for(;e&&(void 0===e.tagName||!s(e)||!e.href);)e=e.parentNode;return e}(e.target),r=t&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||w<r)return!1;if(h(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(n=r)&&(a=n.split(".").pop(),m.some(function(e){return e===a}))?f(e,t,{name:"File Download",props:{url:r}}):void 0}function f(e,t,r){var n,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((n={props:r.props}).revenue=r.revenue,plausible(r.name,n)):((n={props:r.props,callback:i}).revenue=r.revenue,plausible(r.name,n),setTimeout(i,5e3),e.preventDefault())}o.addEventListener("click",n),o.addEventListener("auxclick",n);var a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],v=u.getAttribute("file-types"),d=u.getAttribute("add-file-types"),m=v&&v.split(",")||d&&d.split(",").concat(a)||a;function g(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},r=e&&e.classList;if(r)for(var n=0;n<r.length;n++){var a,i,o=r.item(n),u=o.match(/plausible-event-(.+)(=|--)(.+)/),u=(u&&(a=u[1],i=u[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i),o.match(/plausible-revenue-(.+)(=|--)(.+)/));u&&(a=u[1],i=u[3],t.revenue[a]=i)}return t}var w=3;function b(e){if("auxclick"!==e.type||e.button===c){for(var t,r,n,a,i=e.target,o=0;o<=w&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;s(i)&&(t=i),h(i)&&(r=i),i=i.parentNode}r&&(a=g(r),t?(a.props.url=t.href,f(e,t,a)):((e={}).props=a.props,e.revenue=a.revenue,plausible(a.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,r=e.target,n=g(r);function a(){t||(t=!0,r.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),(e={props:n.props,callback:a}).revenue=n.revenue,plausible(n.name,e))}),o.addEventListener("click",b),o.addEventListener("auxclick",b)}();