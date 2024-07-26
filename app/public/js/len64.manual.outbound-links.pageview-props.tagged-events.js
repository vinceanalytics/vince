!function(){"use strict";var i=window.location,o=window.document,l=o.currentScript,u=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return s("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return s(null,t);try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var r={},e=(r.n=e,r.u=t&&t.u?t.u:i.href,r.d=l.getAttribute("data-domain"),r.r=o.referrer||null,t&&t.meta&&(r.m=JSON.stringify(t.meta)),t&&t.props&&(r.p=t.props),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),n=r.p||{},a=(e.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);n[t]=n[t]||e}),r.p=n,new XMLHttpRequest);a.open("POST",u,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(r)),a.onreadystatechange=function(){4===a.readyState&&t&&t.callback&&t.callback({status:a.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var r=0;r<t.length;r++)e.apply(this,t[r]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var p=1;function n(e){var t,r;if("auxclick"!==e.type||e.button===p)return(t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,r){if(!t||m<r)return!1;if(v(t))return!0;return e(t.parentNode,r+1)}(t,0)&&(r=t)&&r.href&&r.host&&r.host!==i.host?f(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function f(e,t,r){var n,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(n={props:r.props},plausible(r.name,n)):(n={props:r.props,callback:i},plausible(r.name,n),setTimeout(i,5e3),e.preventDefault())}function d(e){var e=v(e)?e:e&&e.parentNode,t={name:null,props:{}},r=e&&e.classList;if(r)for(var n=0;n<r.length;n++){var a,i=r.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(a=i[1],i=i[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i)}return t}o.addEventListener("click",n),o.addEventListener("auxclick",n);var m=3;function a(e){if("auxclick"!==e.type||e.button===p){for(var t,r,n,a,i=e.target,o=0;o<=m&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;c(i)&&(t=i),v(i)&&(r=i),i=i.parentNode}r&&(a=d(r),t?(a.props.url=t.href,f(e,t,a)):((e={}).props=a.props,plausible(a.name,e)))}}function v(e){var t=e&&e.classList;if(t)for(var r=0;r<t.length;r++)if(t.item(r).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,r=e.target,n=d(r);function a(){t||(t=!0,r.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),e={props:n.props,callback:a},plausible(n.name,e))}),o.addEventListener("click",a),o.addEventListener("auxclick",a)}();