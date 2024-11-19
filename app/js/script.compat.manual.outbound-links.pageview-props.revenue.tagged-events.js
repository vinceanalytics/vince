!function(){"use strict";var e,t,i=window.location,o=window.document,u=o.getElementById("plausible"),l=u.getAttribute("data-api")||(e=(e=u).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function s(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function n(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(i.hostname)||"file:"===i.protocol)return s("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return s(null,t);try{if("true"===window.localStorage.plausible_ignore)return s("localStorage flag",t)}catch(e){}var n={},e=(n.n=e,n.u=t&&t.u?t.u:i.href,n.d=u.getAttribute("data-domain"),n.r=o.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),u.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),r=n.p||{},a=(e.forEach(function(e){var t=e.replace("event-",""),e=u.getAttribute(e);r[t]=r[t]||e}),n.p=r,new XMLHttpRequest);a.open("POST",l,!0),a.setRequestHeader("Content-Type","text/plain"),a.send(JSON.stringify(n)),a.onreadystatechange=function(){4===a.readyState&&t&&t.callback&&t.callback({status:a.status})}}var r=window.plausible&&window.plausible.q||[];window.plausible=n;for(var a=0;a<r.length;a++)n.apply(this,r[a]);function p(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var c=1;function f(e){var t,n;if("auxclick"!==e.type||e.button===c)return(t=function(e){for(;e&&(void 0===e.tagName||!p(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,n){if(!t||m<n)return!1;if(b(t))return!0;return e(t.parentNode,n+1)}(t,0)&&(n=t)&&n.href&&n.host&&n.host!==i.host?v(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function v(e,t,n){var r,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((r={props:n.props}).revenue=n.revenue,plausible(n.name,r)):((r={props:n.props,callback:i}).revenue=n.revenue,plausible(n.name,r),setTimeout(i,5e3),e.preventDefault())}function d(e){var e=b(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},n=e&&e.classList;if(n)for(var r=0;r<n.length;r++){var a,i,o=n.item(r),u=o.match(/plausible-event-(.+)(=|--)(.+)/),u=(u&&(a=u[1],i=u[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i),o.match(/plausible-revenue-(.+)(=|--)(.+)/));u&&(a=u[1],i=u[3],t.revenue[a]=i)}return t}o.addEventListener("click",f),o.addEventListener("auxclick",f);var m=3;function g(e){if("auxclick"!==e.type||e.button===c){for(var t,n,r,a,i=e.target,o=0;o<=m&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;p(i)&&(t=i),b(i)&&(n=i),i=i.parentNode}n&&(a=d(n),t?(a.props.url=t.href,v(e,t,a)):((e={}).props=a.props,e.revenue=a.revenue,plausible(a.name,e)))}}function b(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,n=e.target,r=d(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),(e={props:r.props,callback:a}).revenue=r.revenue,plausible(r.name,e))}),o.addEventListener("click",g),o.addEventListener("auxclick",g)}();