!function(){"use strict";var u=window.location,o=window.document,l=o.currentScript,s=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var n=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){n=!n||n.split(",").some(a),r=r&&r.split(",").some(a);if(!n||r)return c("exclusion rule",t)}function a(e){var t=u.pathname;return(t+=u.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var n={},i=(n.n=e,n.u=u.href,n.d=l.getAttribute("data-domain"),n.r=o.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),n.h=1,new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n,r=0;r<t.length;r++)e.apply(this,t[r]);function a(){n=u.pathname,e("pageview")}function p(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",a),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){n||"visible"!==o.visibilityState||a()}):a();var f=1;function i(e){var t,n;if("auxclick"!==e.type||e.button===f)return(t=function(e){for(;e&&(void 0===e.tagName||!p(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,n){if(!t||m<n)return!1;if(h(t))return!0;return e(t.parentNode,n+1)}(t,0)&&(n=t)&&n.href&&n.host&&n.host!==u.host?v(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function v(e,t,n){var r,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((r={props:n.props}).revenue=n.revenue,plausible(n.name,r)):((r={props:n.props,callback:i}).revenue=n.revenue,plausible(n.name,r),setTimeout(i,5e3),e.preventDefault())}function d(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},n=e&&e.classList;if(n)for(var r=0;r<n.length;r++){var a,i,u=n.item(r),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(a=o[1],i=o[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(a=o[1],i=o[3],t.revenue[a]=i)}return t}o.addEventListener("click",i),o.addEventListener("auxclick",i);var m=3;function g(e){if("auxclick"!==e.type||e.button===f){for(var t,n,r,a,i=e.target,u=0;u<=m&&i;u++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;p(i)&&(t=i),h(i)&&(n=i),i=i.parentNode}n&&(a=d(n),t?(a.props.url=t.href,v(e,t,a)):((e={}).props=a.props,e.revenue=a.revenue,plausible(a.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,n=e.target,r=d(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),(e={props:r.props,callback:a}).revenue=r.revenue,plausible(r.name,e))}),o.addEventListener("click",g),o.addEventListener("auxclick",g)}();